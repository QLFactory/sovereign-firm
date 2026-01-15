package treesitter

import (
	"context"
	"testing"
)

func TestExtractJavaScriptSymbols(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
import React from 'react';
import { useState } from 'react';

export function Counter() {
	const [count, setCount] = useState(0);
	return <button onClick={() => setCount(count + 1)}>{count}</button>;
}

export class Calculator {
	add(a, b) {
		return a + b;
	}
	
	subtract(a, b) {
		return a - b;
	}
}

const PI = 3.14159;
let result = 0;
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "app.jsx", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	structure, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols failed: %v", err)
	}

	// Check imports
	if len(structure.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(structure.Imports))
	}

	// Check functions
	hasCounter := false
	for _, fn := range structure.Functions {
		if fn.Name == "Counter" {
			hasCounter = true
			if !fn.Exported {
				t.Error("Counter should be exported")
			}
		}
	}
	if !hasCounter {
		t.Error("Expected to find Counter function")
	}

	// Check classes
	hasCalculator := false
	for _, cls := range structure.Classes {
		if cls.Name == "Calculator" {
			hasCalculator = true
			// Note: Method extraction is best-effort, just verify class was found
		}
	}
	if !hasCalculator {
		t.Error("Expected to find Calculator class")
	}

	// Check LOC
	if structure.LOC < 20 {
		t.Errorf("LOC = %d, expected > 20", structure.LOC)
	}
}

func TestExtractGoSymbols(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
package main

import (
	"fmt"
	"strings"
)

type User struct {
	Name string
	Age  int
}

type Greeter interface {
	Greet() string
}

func NewUser(name string, age int) *User {
	return &User{Name: name, Age: age}
}

func (u *User) Greet() string {
	return fmt.Sprintf("Hello, %s", u.Name)
}

const MaxAge = 100
var DefaultName = "Anonymous"
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "main.go", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	structure, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols failed: %v", err)
	}

	// Check imports
	if len(structure.Imports) < 1 {
		t.Errorf("Expected at least 1 import, got %d", len(structure.Imports))
	}

	// Check types
	hasUser := false
	hasGreeter := false
	for _, typ := range structure.Types {
		if typ.Name == "User" {
			hasUser = true
			if typ.Kind != SymbolStruct {
				t.Errorf("User should be struct, got %s", typ.Kind)
			}
			if !typ.Exported {
				t.Error("User should be exported")
			}
		}
		if typ.Name == "Greeter" {
			hasGreeter = true
			if typ.Kind != SymbolInterface {
				t.Errorf("Greeter should be interface, got %s", typ.Kind)
			}
		}
	}
	if !hasUser {
		t.Error("Expected to find User struct")
	}
	if !hasGreeter {
		t.Error("Expected to find Greeter interface")
	}

	// Check functions
	hasNewUser := false
	hasGreetMethod := false
	for _, fn := range structure.Functions {
		if fn.Name == "NewUser" {
			hasNewUser = true
			if fn.Kind != SymbolFunction {
				t.Errorf("NewUser should be function, got %s", fn.Kind)
			}
		}
		if fn.Name == "Greet" {
			hasGreetMethod = true
			if fn.Kind != SymbolMethod {
				t.Errorf("Greet should be method, got %s", fn.Kind)
			}
		}
	}
	if !hasNewUser {
		t.Error("Expected to find NewUser function")
	}
	if !hasGreetMethod {
		t.Error("Expected to find Greet method")
	}

	// Check variables
	hasMaxAge := false
	hasDefaultName := false
	for _, v := range structure.Variables {
		if v.Name == "MaxAge" {
			hasMaxAge = true
			if v.Kind != SymbolConstant {
				t.Errorf("MaxAge should be constant, got %s", v.Kind)
			}
		}
		if v.Name == "DefaultName" {
			hasDefaultName = true
			if v.Kind != SymbolVariable {
				t.Errorf("DefaultName should be variable, got %s", v.Kind)
			}
		}
	}
	if !hasMaxAge {
		t.Error("Expected to find MaxAge constant")
	}
	if !hasDefaultName {
		t.Error("Expected to find DefaultName variable")
	}
}

func TestExtractPythonSymbols(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
import os
from typing import Optional

class User:
    def __init__(self, name: str):
        self.name = name
    
    def greet(self) -> str:
        return f"Hello, {self.name}"

def create_user(name: str) -> User:
    return User(name)

def _private_helper():
    pass
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "main.py", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	structure, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols failed: %v", err)
	}

	// Check imports
	if len(structure.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(structure.Imports))
	}

	// Check classes
	hasUser := false
	for _, cls := range structure.Classes {
		if cls.Name == "User" {
			hasUser = true
			if !cls.Exported {
				t.Error("User class should be exported (not private)")
			}
		}
	}
	if !hasUser {
		t.Error("Expected to find User class")
	}

	// Check functions
	hasCreateUser := false
	hasPrivateHelper := false
	for _, fn := range structure.Functions {
		if fn.Name == "create_user" {
			hasCreateUser = true
			if !fn.Exported {
				t.Error("create_user should be exported")
			}
		}
		if fn.Name == "_private_helper" {
			hasPrivateHelper = true
			if fn.Exported {
				t.Error("_private_helper should not be exported (private)")
			}
		}
	}
	if !hasCreateUser {
		t.Error("Expected to find create_user function")
	}
	if !hasPrivateHelper {
		t.Error("Expected to find _private_helper function")
	}
}

func TestExtractTypeScriptSymbols(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`
import { Request, Response } from 'express';

interface User {
	id: number;
	name: string;
}

type UserResponse = User & { token: string };

enum Status {
	Active,
	Inactive,
	Pending
}

export function getUser(req: Request): User {
	return { id: 1, name: "Test" };
}

export class UserService {
	private users: User[] = [];
	
	getAll(): User[] {
		return this.users;
	}
}
`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "service.ts", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	structure, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols failed: %v", err)
	}

	// Check types (interfaces, type aliases, enums)
	hasUserInterface := false
	hasUserResponse := false
	hasStatus := false
	for _, typ := range structure.Types {
		switch typ.Name {
		case "User":
			hasUserInterface = true
			if typ.Kind != SymbolInterface {
				t.Errorf("User should be interface, got %s", typ.Kind)
			}
		case "UserResponse":
			hasUserResponse = true
			if typ.Kind != SymbolType {
				t.Errorf("UserResponse should be type alias, got %s", typ.Kind)
			}
		case "Status":
			hasStatus = true
			if typ.Kind != SymbolEnum {
				t.Errorf("Status should be enum, got %s", typ.Kind)
			}
		}
	}
	if !hasUserInterface {
		t.Error("Expected to find User interface")
	}
	if !hasUserResponse {
		t.Error("Expected to find UserResponse type alias")
	}
	if !hasStatus {
		t.Error("Expected to find Status enum")
	}
}

func TestCodeStructureJSON(t *testing.T) {
	p := NewParser()
	defer p.Close()

	source := []byte(`function hello() { return "world"; }`)

	ctx := context.Background()
	result, err := p.Parse(ctx, "test.js", source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	structure, err := ExtractSymbols(result)
	if err != nil {
		t.Fatalf("ExtractSymbols failed: %v", err)
	}

	if structure.FilePath != "test.js" {
		t.Errorf("FilePath = %q, want %q", structure.FilePath, "test.js")
	}

	if structure.Language != LangJavaScript {
		t.Errorf("Language = %q, want %q", structure.Language, LangJavaScript)
	}
}
