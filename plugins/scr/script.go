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
	"git.golaxy.org/core/service"
	"git.golaxy.org/core/utils/option"
	"git.golaxy.org/framework/plugins/log"
	"git.golaxy.org/scaffold/plugins/scr/fwlib"
	"github.com/fsnotify/fsnotify"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"io/fs"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"time"
)

// IScript 脚本插件接口
type IScript interface {
	// Hotfix 热更新
	Hotfix() error
	// Eval 执行脚本
	Eval(src string) (reflect.Value, error)
}

func newScript(setting ...option.Setting[ScriptOptions]) IScript {
	return &_Script{
		options: option.Make(With.Default(), setting...),
	}
}

type _Script struct {
	svcCtx    service.Context
	options   ScriptOptions
	intp      *interp.Interpreter
	reloading atomic.Int64
}

// InitSP 初始化服务插件
func (s *_Script) InitSP(svcCtx service.Context) {
	s.svcCtx = svcCtx

	intp, err := s.load()
	if err != nil {
		log.Panicf(s.svcCtx, "init script load %+v failed, %s", s.options.PathList, err)
	}
	s.intp = intp

	log.Infof(s.svcCtx, "init script load %+v ok", s.options.PathList)

	if s.options.AutoHotFix {
		s.autoHotFix()
	}
}

// ShutSP 关闭服务插件
func (s *_Script) ShutSP(svcCtx service.Context) {
	log.Infof(svcCtx, "shut plugin %q", self.Name)
}

// Hotfix 热更新
func (s *_Script) Hotfix() error {
	intp, err := s.load()
	if err != nil {
		return err
	}
	s.intp = intp

	return nil
}

// Eval 执行脚本
func (s *_Script) Eval(src string) (reflect.Value, error) {
	return s.intp.Eval(src)
}

func (s *_Script) load() (*interp.Interpreter, error) {
	intp := interp.New(interp.Options{})
	intp.Use(stdlib.Symbols)
	intp.Use(fwlib.Symbols)

	for _, symbols := range s.options.SymbolsList {
		intp.Use(symbols)
	}

	if err := s.options.LoadingCB.Invoke(func(error) bool { return true }, intp); err != nil {
		return nil, err
	}

	for _, path := range s.options.PathList {
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				return nil
			}

			if _, err := intp.EvalPath(path); err != nil {
				return fmt.Errorf("load script path %q failed, %s", path, err)
			}

			if _, err := intp.Eval(fmt.Sprintf(`import "%s"`, path)); err != nil {
				return fmt.Errorf("import script path %q failed, %s", path, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if err := s.options.LoadedCB.Invoke(func(error) bool { return true }, intp); err != nil {
		return nil, err
	}

	return intp, nil
}

func (s *_Script) autoHotFix() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Panicf(s.svcCtx, "auto hotfix script watch %+v failed, %s", s.options.PathList, err)
	}

	for _, path := range s.options.PathList {
		if err = watcher.AddWith(path); err != nil {
			log.Panicf(s.svcCtx, "auto hotfix script watch %q failed, %s", path, err)
		}
	}

	go func() {
		for {
			select {
			case e, ok := <-watcher.Events:
				if !ok {
					return
				}

				log.Infof(s.svcCtx, "auto hotfix script detecting %q changes, preparing to reload in 10s", e)

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

					intp, err := s.load()
					if err != nil {
						log.Errorf(s.svcCtx, "auto hotfix script reload %+v failed, %s", s.options.PathList, err)
						return
					}
					s.intp = intp

					log.Infof(s.svcCtx, "auto hotfix script reload %+v ok", s.options.PathList)
				}()

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Errorf(s.svcCtx, "auto hotfix script watch %+v failed, %s", s.options.PathList, err)
			}
		}
	}()

	log.Infof(s.svcCtx, "auto hotfix script watch %+v ok", s.options.PathList)
}
