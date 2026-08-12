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

package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	pbIdentifierRegexp     = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	yamlAliasInvalidRegexp = regexp.MustCompile("[\\p{Cc} \\-?:,\\[\\]{}#&*!|>'\"%@`\\\\]")
)

func validatePbIdentifier(name string) error {
	if !pbIdentifierRegexp.MatchString(name) {
		return fmt.Errorf("must match [A-Za-z][A-Za-z0-9_]*")
	}

	return nil
}

func validateYAMLAlias(alias string) error {
	invalid := yamlAliasInvalidRegexp.FindString(alias)
	if invalid != "" {
		return fmt.Errorf("contains reserved character %q", []rune(invalid)[0])
	}

	return nil
}

func snake2Camel(s string) string {
	var buf bytes.Buffer
	upper := true
	for _, c := range s {
		if c == '_' {
			upper = true
			continue
		}
		if upper {
			buf.WriteRune(unicode.ToUpper(c))
			upper = false
		} else {
			buf.WriteRune(c)
		}
	}
	return strings.Title(buf.String())
}
