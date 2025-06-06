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
	"git.golaxy.org/core/ec"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/scaffold/addins/goscr/dynamic"
	"github.com/elliotchance/pie/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/pangdogs/yaegi/stdlib"
	"strings"
	"sync/atomic"
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
		options: option.Make(With.Default(), setting...),
	}
}

type _Script struct {
	svcCtx    service.Context
	options   ScriptOptions
	solution  *dynamic.Solution
	reloading atomic.Int64
}

// Init 初始化插件
func (s *_Script) Init(svcCtx service.Context) {
	s.svcCtx = svcCtx

	solution, err := s.loadSolution()
	if err != nil {
		log.Panicf(s.svcCtx, "init load solution %q failed, %s", s.options.PkgRoot, err)
	}
	s.solution = solution

	log.Infof(s.svcCtx, "init load solution %q ok, projects: [%s]", s.options.PkgRoot,
		strings.Join(pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
			return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
		}), ", "))

	if s.options.AutoHotFix {
		s.autoHotFix()
	}
}

// Shut 关闭插件
func (s *_Script) Shut(svcCtx service.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)
}

// Solution 解决方案
func (s *_Script) Solution() *dynamic.Solution {
	return s.solution
}

// Hotfix 热更新
func (s *_Script) Hotfix() error {
	solution, err := s.loadSolution()
	if err != nil {
		return err
	}
	s.solution = solution

	return nil
}

func (s *_Script) OnServiceRunningStatusChanged(svcCtx service.Context, status service.RunningStatus, args ...any) {
	switch status {
	case service.RunningStatus_EntityPTDeclared, service.RunningStatus_EntityPTRedeclared:
		solution := s.solution
		if solution == nil {
			return
		}
		s.cacheCallPath(solution, args[0].(ec.EntityPT))
	}
}

func (s *_Script) loadSolution() (*dynamic.Solution, error) {
	solution := dynamic.NewSolution(s.options.PkgRoot)
	solution.Use(stdlib.Symbols)

	if err := s.options.LoadingCB.SafeCall(solution); err != nil {
		return nil, fmt.Errorf("loading callback error occurred, %s", err)
	}

	for _, project := range s.options.Projects {
		if err := solution.Load(project); err != nil {
			return nil, fmt.Errorf("load project %q -> %q + %q failed, %s", project.ScriptRoot, project.LocalPath, project.RemoteURL, err)
		}
	}

	if err := s.options.LoadedCB.SafeCall(solution); err != nil {
		return nil, fmt.Errorf("loaded callback error occurred, %s", err)
	}

	s.svcCtx.GetEntityLib().Range(func(entityPT ec.EntityPT) bool {
		s.cacheCallPath(solution, entityPT)
		return true
	})

	return solution, nil
}

func (s *_Script) autoHotFix() {
	if pie.Any(s.options.Projects, func(project *dynamic.Project) bool { return project.LocalPath != "" }) {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Panicf(s.svcCtx, "auto hotfix solution %q watch local changes failed, projects: [%s], %s",
				s.options.PkgRoot,
				strings.Join(pie.Of(s.options.Projects).Filter(func(project *dynamic.Project) bool {
					return project.LocalPath != ""
				}).StringsUsing(func(project *dynamic.Project) string {
					return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
				}), ", "),
				err)
		}

		for _, project := range s.options.Projects {
			if project.LocalPath != "" {
				if err = watcher.Add(project.LocalPath); err != nil {
					watcher.Close()
					log.Panicf(s.svcCtx, "auto hotfix solution %q watch %q -> %q + %q local changes failed, %s", s.options.PkgRoot, project.ScriptRoot, project.LocalPath, project.RemoteURL, err)
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

					log.Infof(s.svcCtx, "auto hotfix solution %q detecting local %q %s changes, preparing to reload in %s", s.options.PkgRoot, e.Name, e.Op, s.options.AutoHotFixLocalDetectingDelayTime)

					s.reloading.Add(1)

					go func() {
						time.Sleep(s.options.AutoHotFixLocalDetectingDelayTime)

						if s.reloading.Add(-1) != 0 {
							return
						}

						select {
						case <-s.svcCtx.Done():
							return
						default:
						}

						if err := s.Hotfix(); err != nil {
							log.Errorf(s.svcCtx, "auto hotfix load solution %q failed, %s", s.options.PkgRoot, err)
							return
						}

						log.Infof(s.svcCtx, "auto hotfix load solution %q ok, projects: [%s]", s.options.PkgRoot,
							strings.Join(pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
								return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
							}), ", "))
					}()

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Errorf(s.svcCtx, "auto hotfix solution %q watch local changes failed, projects: [%s], %s",
						s.options.PkgRoot,
						strings.Join(pie.Of(s.options.Projects).Filter(func(project *dynamic.Project) bool {
							return project.LocalPath != ""
						}).StringsUsing(func(project *dynamic.Project) string {
							return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
						}), ", "),
						err)
				}
			}
		}()

		log.Infof(s.svcCtx, "auto hotfix solution %q watch local changes ok, projects: [%s]",
			s.options.PkgRoot,
			strings.Join(pie.Of(s.options.Projects).Filter(func(project *dynamic.Project) bool {
				return project.LocalPath != ""
			}).StringsUsing(func(project *dynamic.Project) string {
				return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
			}), ", "))
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
					log.Panicf(s.svcCtx, "auto hotfix solution %q watch remote changes failed, projects: [%s], %s",
						s.options.PkgRoot,
						strings.Join(pie.Of(s.options.Projects).Filter(func(project *dynamic.Project) bool {
							return project.RemoteURL != ""
						}).StringsUsing(func(project *dynamic.Project) string {
							return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
						}), ", "),
						err)
					continue
				}
				if !b {
					continue
				}

				if err := s.Hotfix(); err != nil {
					log.Errorf(s.svcCtx, "auto hotfix load solution %q failed, %s", s.options.PkgRoot, err)
					continue
				}

				log.Infof(s.svcCtx, "auto hotfix load solution %q ok, projects: [%s]", s.options.PkgRoot,
					strings.Join(pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
						return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
					}), ", "))
			}
		}()

		log.Infof(s.svcCtx, "auto hotfix solution %q watch remote changes ok, projects: [%s]",
			s.options.PkgRoot,
			strings.Join(pie.Of(s.options.Projects).Filter(func(project *dynamic.Project) bool {
				return project.RemoteURL != ""
			}).StringsUsing(func(project *dynamic.Project) string {
				return fmt.Sprintf("%q -> %q + %q", project.ScriptRoot, project.LocalPath, project.RemoteURL)
			}), ", "))
	}
}

func (s *_Script) cacheCallPath(solution *dynamic.Solution, entityPT ec.EntityPT) {
	scriptPkg, ok := entityPT.Extra().Get("script_pkg")
	if ok {
		scriptIdent, ok := entityPT.Extra().Get("script_ident")
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
		comp := entityPT.Component(i)

		scriptPkg, ok := comp.Extra.Get("script_pkg")
		if ok {
			scriptIdent, ok := comp.Extra.Get("script_ident")
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
