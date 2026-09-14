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
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const metaQueryEncodingHint = "Meta values use URL query encoding; common replacements are '+' as '%2B', '&' as '%26', '%' as '%25', and ';' as '%3B'"

var metaStructValidator = newMetaStructValidator()

func newMetaStructValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("form"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	if err := validate.RegisterValidation("valid_separator", func(field validator.FieldLevel) bool {
		return validateSeparator(field.Field().String()) == nil
	}); err != nil {
		log.Panic(err)
	}
	if err := validate.RegisterValidation("valid_scope", func(field validator.FieldLevel) bool {
		return validateScope(field.Field().String()) == nil
	}); err != nil {
		log.Panic(err)
	}
	if err := validate.RegisterValidation("valid_pb_field_number", func(field validator.FieldLevel) bool {
		return checkPbFieldNumber(int32(field.Field().Int())) == nil
	}); err != nil {
		log.Panic(err)
	}

	return validate
}

func validateMeta(meta *Meta) error {
	err := metaStructValidator.Struct(meta)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return err
	}

	fieldError := validationErrors[0]
	parameter := fieldError.Field()

	switch fieldError.Tag() {
	case "required":
		return fmt.Errorf("meta parameter %q cannot be empty", parameter)
	case "gte":
		return fmt.Errorf("meta parameter %q value %v must be greater than or equal to %s", parameter, fieldError.Value(), fieldError.Param())
	case "valid_separator":
		return fmt.Errorf("invalid separator: %w; %s", validateSeparator(meta.Separator), metaQueryEncodingHint)
	case "valid_scope":
		return fmt.Errorf("invalid scope %q: %w", fieldError.Value(), validateScope(fieldError.Value().(string)))
	case "valid_pb_field_number":
		return checkPbFieldNumber(*meta.PbFieldNumber)
	default:
		return fmt.Errorf("invalid meta parameter %q: %w", parameter, err)
	}
}
