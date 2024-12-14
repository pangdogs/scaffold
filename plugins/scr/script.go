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

package scr

import (
	"fmt"
	"git.golaxy.org/core/runtime"
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/scaffold/plugins/scr/dynamic"
	"git.golaxy.org/scaffold/plugins/scr/fwlib"
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

	log.Infof(s.svcCtx, "init load solution %q ok, projects: %q", s.options.PkgRoot,
		pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
			return fmt.Sprintf("%s - %s", project.PkgRoot, project.LocalPath)
		}))

	if s.options.AutoHotFix {
		s.autoHotFix()
	}
}

// Shut 关闭插件
func (s *_Script) Shut(svcCtx service.Context, _ runtime.Context) {
	log.Infof(svcCtx, "shut plugin %q", self.Name)
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

func (s *_Script) loadSolution() (*dynamic.Solution, error) {
	solution := dynamic.NewSolution(s.options.PkgRoot)
	solution.Use(stdlib.Symbols)
	solution.Use(fwlib.Symbols)

	if err := s.options.LoadingCB.Invoke(func(err error) bool { return err != nil }, solution); err != nil {
		return nil, fmt.Errorf("loading callback error occurred, %s", err)
	}

	for _, project := range s.options.Projects {
		if err := solution.Load(project); err != nil {
			return nil, fmt.Errorf("load project %q failed, %s", fmt.Sprintf("%s - %s", project.PkgRoot, project.LocalPath), err)
		}
	}

	if err := s.options.LoadedCB.Invoke(func(err error) bool { return err != nil }, solution); err != nil {
		return nil, fmt.Errorf("loaded callback error occurred, %s", err)
	}

	return solution, nil
}

func (s *_Script) autoHotFix() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panicf(s.svcCtx, "auto hotfix solution %q watch changes failed, projects: %q, %s",
			s.options.PkgRoot,
			pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
				return fmt.Sprintf("%s - %s", project.PkgRoot, project.LocalPath)
			}),
			err)
	}

	for _, project := range s.options.Projects {
		if err = watcher.Add(project.LocalPath); err != nil {
			log.Panicf(s.svcCtx, "auto hotfix solution %q watch changes failed, projects: %q, %s",
				s.options.PkgRoot,
				pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
					return fmt.Sprintf("%s - %s", project.PkgRoot, project.LocalPath)
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

					log.Infof(s.svcCtx, "auto hotfix load solution %q ok, projects: %q", s.options.PkgRoot,
						pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
							return fmt.Sprintf("%s - %s", project.PkgRoot, project.LocalPath)
						}))
				}()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Errorf(s.svcCtx, "auto hotfix solution %q watch changes failed, projects: %q, %s",
					s.options.PkgRoot,
					pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
						return fmt.Sprintf("%s - %s", project.PkgRoot, project.LocalPath)
					}),
					err)
			}
		}
	}()

	log.Panicf(s.svcCtx, "auto hotfix solution %q watch changes ok, projects: %q",
		s.options.PkgRoot,
		pie.Of(s.options.Projects).StringsUsing(func(project *dynamic.Project) string {
			return fmt.Sprintf("%s - %s", project.PkgRoot, project.LocalPath)
		}))
}
