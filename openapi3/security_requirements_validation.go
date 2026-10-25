package openapi3

import (
	"context"
	"fmt"
	"strconv"
)

// validateSecurityRequirements resolves implicit names against the entry
// document, including names in resolved external paths and callbacks. OpenAPI
// 3.2 also permits URI references, so this check applies only to 3.0 and 3.1.
func (doc *T) validateSecurityRequirements(ctx context.Context) error {
	if version := doc.OpenAPIMajorMinor(); version != "3.0" && version != "3.1" {
		return nil
	}
	v := securityRequirementsValidator{
		ctx: ctx, paths: make(map[*PathItem]bool),
		operations: make(map[*Operation]bool), callbacks: make(map[*Callback]bool),
	}
	if doc.Components != nil {
		v.schemes = doc.Components.SecuritySchemes
	}
	me := newErrCollector(ctx)
	section := func(name string) func(error) error {
		return func(err error) error { return &SectionValidationError{Section: name, Cause: err} }
	}
	if err := me.emitWrapped(section("security"), v.requirements("/security", doc.Security, doc.Origin)); err != nil {
		return err
	}
	if doc.Paths != nil {
		for _, name := range componentNames(doc.Paths.Map()) {
			wrap := func(err error) error {
				return section("paths")(&PathValidationError{Path: name, Cause: err})
			}
			if err := me.emitWrapped(wrap, v.pathItem("/paths/"+escapeRefString(name), doc.Paths.Value(name))); err != nil {
				return err
			}
		}
	}
	for _, name := range componentNames(doc.Webhooks) {
		wrap := func(err error) error {
			return section("webhooks")(&WebhookValidationError{Name: name, Cause: err})
		}
		if err := me.emitWrapped(wrap, v.pathItem("/webhooks/"+escapeRefString(name), doc.Webhooks[name])); err != nil {
			return err
		}
	}
	if doc.Components != nil {
		for _, name := range componentNames(doc.Components.Callbacks) {
			wrap := func(err error) error {
				return section("components")(&ComponentValidationError{Section: "callback", Name: name, Cause: err})
			}
			if err := me.emitWrapped(wrap, v.callback("/components/callbacks/"+escapeRefString(name), doc.Components.Callbacks[name])); err != nil {
				return err
			}
		}
	}
	return me.result()
}

type securityRequirementsValidator struct {
	ctx        context.Context
	schemes    SecuritySchemes
	paths      map[*PathItem]bool
	operations map[*Operation]bool
	callbacks  map[*Callback]bool
}

func (v *securityRequirementsValidator) requirements(ptr string, requirements SecurityRequirements, origin *Origin) error {
	me := newErrCollector(v.ctx)
	for i, requirement := range requirements {
		for _, name := range componentNames(requirement) {
			if _, declared := v.schemes[name]; declared {
				continue
			}
			pointer := ptr + "/" + strconv.Itoa(i) + "/" + escapeRefString(name)
			err := &SecurityRequirementSchemeUndefinedError{
				ValidationError: ValidationError{Message: fmt.Sprintf("security requirement %s refers to undefined security scheme %q", pointer, name)},
				Scheme:          name, JSONPointer: pointer, Origin: origin,
			}
			if err := me.emit(err); err != nil {
				return err
			}
		}
	}
	return me.result()
}

func (v *securityRequirementsValidator) pathItem(ptr string, item *PathItem) error {
	if item == nil || v.paths[item] {
		return nil
	}
	v.paths[item] = true
	me := newErrCollector(v.ctx)
	for _, entry := range item.operationEntries() {
		wrap := func(err error) error { return &OperationValidationError{Method: entry.method, Cause: err} }
		if err := me.emitWrapped(wrap, v.operation(ptr+entry.pointerSuffix, entry.operation)); err != nil {
			return err
		}
	}
	return me.result()
}

func (v *securityRequirementsValidator) operation(ptr string, operation *Operation) error {
	if v.operations[operation] {
		return nil
	}
	v.operations[operation] = true
	me := newErrCollector(v.ctx)
	if operation.Security != nil {
		if err := me.emit(v.requirements(ptr+"/security", *operation.Security, operation.Origin)); err != nil {
			return err
		}
	}
	for _, name := range componentNames(operation.Callbacks) {
		if err := me.emit(v.callback(ptr+"/callbacks/"+escapeRefString(name), operation.Callbacks[name])); err != nil {
			return err
		}
	}
	return me.result()
}

func (v *securityRequirementsValidator) callback(ptr string, ref *CallbackRef) error {
	if ref == nil || ref.Value == nil || v.callbacks[ref.Value] {
		return nil
	}
	v.callbacks[ref.Value] = true
	me := newErrCollector(v.ctx)
	for _, expression := range componentNames(ref.Value.Map()) {
		if err := me.emit(v.pathItem(ptr+"/"+escapeRefString(expression), ref.Value.Value(expression))); err != nil {
			return err
		}
	}
	return me.result()
}
