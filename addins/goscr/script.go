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
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/addins/log"
	"git.golaxy.org/framework/addins/rpc/callpath"
	"git.golaxy.org/scaffold/addins/goscr/dynamic"
	"github.com/elliotchance/pie/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/pangdogs/yaegi/stdlib"
	"sync/atomic"
	"time"
)

// IScript 脚本插件接口
type IScript interface {
	// Hotfix 热更新
	Hotfix() error
	// Solution 解决方案
	Solution() *dynamic.Solution
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
func (s *_Script) Init(svcCtx service.Context, _ runtime.Context) {
	s.svcCtx = svcCtx

	solution, err := s.loadSolution()
	if err != nil {
		log.Panicf(s.svcCtx, "init load solution %q failed, %s", s.options.PkgRoot, err)
	}
	s.solution = solution

	log.Infof(s.svcCtx, "init load solution %q ok, projects: %s", s.options.PkgRoot,
		pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
			return fmt.Sprintf("%q -> %q", project.PkgRoot, project.LocalPath)
		}))

	if s.options.AutoHotFix {
		s.autoHotFix()
	}
}

// Shut 关闭插件
func (s *_Script) Shut(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "shut addin %q", self.Name)
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

// Solution 解决方案
func (s *_Script) Solution() *dynamic.Solution {
	return s.solution
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
			return nil, fmt.Errorf("load project %s failed, %s", fmt.Sprintf("%q -> %q", project.PkgRoot, project.LocalPath), err)
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
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panicf(s.svcCtx, "auto hotfix solution %q watch changes failed, projects: %s, %s",
			s.options.PkgRoot,
			pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
				return fmt.Sprintf("%q -> %q", project.PkgRoot, project.LocalPath)
			}),
			err)
	}

	for _, project := range s.options.Projects {
		if err = watcher.Add(project.LocalPath); err != nil {
			log.Panicf(s.svcCtx, "auto hotfix solution %q watch changes failed, projects: %s, %s",
				s.options.PkgRoot,
				pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
					return fmt.Sprintf("%q -> %q", project.PkgRoot, project.LocalPath)
				}),
				err)
		}
	}

	go func() {
		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					return
				}

				log.Infof(s.svcCtx, "auto hotfix solution %q detecting %q changes, preparing to reload in 10s", s.options.PkgRoot, e)

				s.reloading.Add(1)

				go func() {
					time.Sleep(3 * time.Second)

					if s.reloading.Add(-1) != 0 {
						return
					}

					select {
					case <-s.svcCtx.Done():
						return
					default:
					}

					solution, err := s.loadSolution()
					if err != nil {
						log.Errorf(s.svcCtx, "auto hotfix load solution %q failed, %s", s.options.PkgRoot, err)
						return
					}
					s.solution = solution

					log.Infof(s.svcCtx, "auto hotfix load solution %q ok, projects: %s", s.options.PkgRoot,
						pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
							return fmt.Sprintf("%q -> %q", project.PkgRoot, project.LocalPath)
						}))
				}()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Errorf(s.svcCtx, "auto hotfix solution %q watch changes failed, projects: %s, %s",
					s.options.PkgRoot,
					pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
						return fmt.Sprintf("%q -> %q", project.PkgRoot, project.LocalPath)
					}),
					err)
			}
		}
	}()

	log.Infof(s.svcCtx, "auto hotfix solution %q watch changes ok, projects: %s",
		s.options.PkgRoot,
		pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
			return fmt.Sprintf("%q -> %q", project.PkgRoot, project.LocalPath)
		}))
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
