/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package goscr

import (
	"fmt"
	"sync"

	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/scaffold/addins/goscr/dynamic"
	"github.com/elliotchance/pie/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/pangdogs/yaegi/stdlib"
	"go.uber.org/zap"

	"time"
)

// IScript 脚本插件接口
type IScript interface {
	// Solution 解决方案
	Solution() *dynamic.Solution
	// Hotfix 热更新
	Hotfix() error
}

func newScript(setting ...option.Setting[ScriptOptions]) IScript {
	return &_Script{
		options: option.New(With.Default(), setting...),
	}
}

type _Script struct {
	svcCtx      service.Context
	options     ScriptOptions
	solution    *dynamic.Solution
	reloadingMu sync.Mutex
}

// Init 初始化插件
func (s *_Script) Init(svcCtx service.Context) {
	log.L(svcCtx).Info("initializing add-in", zap.String("name", AddIn.Name))

	s.svcCtx = svcCtx

	solution, err := s.loadSolution()
	if err != nil {
		log.L(s.svcCtx).Panic("init load solution failed",
			zap.String("pkg_root", s.options.PkgRoot),
			zap.Error(err))
	}
	s.solution = solution

	log.L(s.svcCtx).Info("init load solution ok",
		zap.String("pkg_root", s.options.PkgRoot),
		zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)))

	if s.options.AutoHotFix {
		s.autoHotFix()
	}
}

// Shut 关闭插件
func (s *_Script) Shut(svcCtx service.Context) {
	log.L(svcCtx).Info("shutting down add-in", zap.String("name", AddIn.Name))
}

// Solution 解决方案
func (s *_Script) Solution() *dynamic.Solution {
	return s.solution
}

// Hotfix 热更新
func (s *_Script) Hotfix() error {
	solution, err := s.loadSolution()
	if err != nil {
		log.L(s.svcCtx).Error("hotfix load solution failed",
			zap.String("pkg_root", s.options.PkgRoot),
			zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)),
			zap.Error(err))
		return err
	}
	s.solution = solution

	log.L(s.svcCtx).Info("hotfix load solution ok",
		zap.String("pkg_root", s.options.PkgRoot),
		zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)))
	return nil
}

func (s *_Script) loadSolution() (*dynamic.Solution, error) {
	solution := dynamic.NewSolution(s.options.PkgRoot)
	solution.Use(stdlib.Symbols)

	if err := s.options.LoadingCB.SafeCall(solution); err != nil {
		return nil, fmt.Errorf("loading callback error occurred, %s", err)
	}

	for _, project := range s.options.Projects {
		if err := solution.Load(project); err != nil {
			return nil, fmt.Errorf("load project failed, project:%s, %s", s.showProject(project), err)
		}
	}

	if err := s.options.LoadedCB.SafeCall(solution); err != nil {
		return nil, fmt.Errorf("loaded callback error occurred, %s", err)
	}

	for _, entityPT := range s.svcCtx.EntityLib().List() {
		s.cacheCallPath(solution, entityPT)
	}

	return solution, nil
}

func (s *_Script) autoHotFix() {
	if pie.Any(s.options.Projects, func(project *dynamic.Project) bool { return project.LocalPath != "" }) {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.L(s.svcCtx).Panic("auto hotfix watch solution local path changes failed",
				zap.String("pkg_root", s.options.PkgRoot),
				zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)),
				zap.Error(err))
		}

		for _, project := range s.options.Projects {
			if project.LocalPath != "" {
				if err = watcher.Add(project.LocalPath); err != nil {
					watcher.Close()
					log.L(s.svcCtx).Panic("auto hotfix watch project local path changes failed",
						zap.String("pkg_root", s.options.PkgRoot),
						zap.String("project", s.showProject(project)),
						zap.Error(err))
				}
			}
		}

		go func() {
			defer watcher.Close()
			for {
				select {
				case <-s.svcCtx.Done():
					return
				case e, ok := <-watcher.Events:
					if !ok {
						return
					}

					log.L(s.svcCtx).Info("auto hotfix detecting solution local path changes, preparing to reload in delay_time",
						zap.String("pkg_root", s.options.PkgRoot),
						zap.String("file_path", e.Name),
						zap.String("file_op", e.Op.String()),
						zap.Duration("delay_time", s.options.AutoHotFixLocalDetectingDelayTime))

					go func() {
						if !s.reloadingMu.TryLock() {
							return
						}
						defer s.reloadingMu.Unlock()

						time.Sleep(s.options.AutoHotFixLocalDetectingDelayTime)

						select {
						case <-s.svcCtx.Done():
							return
						default:
						}

						solution, err := s.loadSolution()
						if err != nil {
							log.L(s.svcCtx).Error("auto hotfix load solution failed",
								zap.String("pkg_root", s.options.PkgRoot),
								zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)),
								zap.Error(err))
							return
						}
						s.solution = solution

						log.L(s.svcCtx).Info("auto hotfix load solution ok",
							zap.String("pkg_root", s.options.PkgRoot),
							zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)))
					}()

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.L(s.svcCtx).Error("auto hotfix watch solution local path changes failed",
						zap.String("pkg_root", s.options.PkgRoot),
						zap.Error(err))
				}
			}
		}()

		log.L(s.svcCtx).Info("auto hotfix watch solution local path changes ok",
			zap.String("pkg_root", s.options.PkgRoot),
			zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)))
	}

	if pie.Any(s.options.Projects, func(project *dynamic.Project) bool { return project.RemoteURL != "" }) {
		go func() {
			for {
				time.Sleep(s.options.AutoHotFixRemoteCheckingIntervalTime)

				select {
				case <-s.svcCtx.Done():
					return
				default:
				}

				b, err := s.solution.DetectRemoteChanged()
				if err != nil {
					log.L(s.svcCtx).Error("auto hotfix detect solution remote path changes failed",
						zap.String("pkg_root", s.options.PkgRoot),
						zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)),
						zap.Error(err))
					continue
				}
				if !b {
					continue
				}

				log.L(s.svcCtx).Info("auto hotfix detecting solution remote path changes, preparing to reload",
					zap.String("pkg_root", s.options.PkgRoot))

				func() {
					if !s.reloadingMu.TryLock() {
						return
					}
					defer s.reloadingMu.Unlock()

					solution, err := s.loadSolution()
					if err != nil {
						log.L(s.svcCtx).Error("auto hotfix load solution failed",
							zap.String("pkg_root", s.options.PkgRoot),
							zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)),
							zap.Error(err))
						return
					}
					s.solution = solution

					log.L(s.svcCtx).Info("auto hotfix load solution ok",
						zap.String("pkg_root", s.options.PkgRoot),
						zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)))
				}()
			}
		}()

		log.L(s.svcCtx).Info("auto hotfix watch solution remote path changes ok",
			zap.String("pkg_root", s.options.PkgRoot),
			zap.Strings("projects", pie.Of(s.options.Projects).StringsUsing(s.showProject)))
	}
}

func (s *_Script) cacheCallPath(solution *dynamic.Solution, entityPT ec.EntityPT) {
	scriptPkg, ok := entityPT.Meta().Get("script_pkg")
	if ok {
		scriptIdent, ok := entityPT.Meta().Get("script_ident")
		if ok {
			script := solution.Package(scriptPkg.(string)).Ident(scriptIdent.(string))
			if script != nil {
				for _, method := range script.Methods {
					callpath.Cache("", method.Name)
				}
			}
		}
	}

	for i := range entityPT.CountComponents() {
		comp := entityPT.GetComponent(i)

		scriptPkg, ok := comp.Meta.Get("script_pkg")
		if ok {
			scriptIdent, ok := comp.Meta.Get("script_ident")
			if ok {
				script := solution.Package(scriptPkg.(string)).Ident(scriptIdent.(string))
				if script != nil {
					for _, method := range script.Methods {
						callpath.Cache(comp.Name, method.Name)
					}
				}
			}
		}
	}
}

func (s *_Script) showProject(project *dynamic.Project) string {
	var files []string
	if project.LocalPath != "" {
		files = append(files, project.LocalPath)
	}
	if project.RemoteURL != "" {
		files = append(files, project.RemoteURL)
	}
	return fmt.Sprintf("%s => %s", project.ScriptRoot, files)
}
