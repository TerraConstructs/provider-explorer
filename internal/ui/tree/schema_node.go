// Copyright (c) 2024 Terraform Constructs
// Licensed under the Apache License, Version 2.0

package tree

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    tfjson "github.com/hashicorp/terraform-json"
    "github.com/zclconf/go-cty/cty"
)

var (
	schemaArgumentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("118"))

	schemaAttributeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("75"))

	schemaRequiredStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	schemaOptionalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("178"))

	schemaComputedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("140"))
)

// SchemaNode represents a node in the schema tree that implements VisibleModel
type SchemaNode struct {
	id          string
	name        string
	displayText string
	path        []string
	nodeType    SchemaNodeType
	visible     bool
	
	// For attributes
	attribute *tfjson.SchemaAttribute
	
	// For blocks  
	block *tfjson.SchemaBlock
}

type SchemaNodeType int

const (
	AttributeNode SchemaNodeType = iota
	BlockNode
)

// NewAttributeNode creates a new schema node for an attribute
func NewAttributeNode(id, name string, attr *tfjson.SchemaAttribute, path []string) *SchemaNode {
    // Create display text with type and status
    typeInfo := ""
    if attr.AttributeType != cty.NilType {
        typeInfo = fmt.Sprintf(" (%s)", typeFriendlyLabel(attr.AttributeType))
    }

	status := ""
	var style lipgloss.Style
	if attr.Required {
		status = " [required]"
		style = schemaRequiredStyle
	} else if attr.Optional {
		status = " [optional]"
		style = schemaOptionalStyle
	} else if attr.Computed {
		status = " [computed]"
		style = schemaComputedStyle
	}

	displayText := style.Render(name + typeInfo + status)

	return &SchemaNode{
		id:          id,
		name:        name,
		displayText: displayText,
		path:        path,
		nodeType:    AttributeNode,
		visible:     true,
		attribute:   attr,
	}
}

// typeFriendlyLabel provides a simple, stable label for cty types for display.
func typeFriendlyLabel(t cty.Type) string {
    switch t { // handle primitives exactly
    case cty.String:
        return "string"
    case cty.Number:
        return "number"
    case cty.Bool:
        return "bool"
    }
    // collections and other shapes (avoid verbose/unstable formatting)
    if t.IsListType() {
        return "list"
    }
    if t.IsSetType() {
        return "set"
    }
    if t.IsMapType() {
        return "map"
    }
    if t.IsTupleType() {
        return "tuple"
    }
    if t.IsObjectType() {
        return "object"
    }
    return "any"
}

// NewBlockNode creates a new schema node for a block
func NewBlockNode(id, name string, block *tfjson.SchemaBlock, path []string) *SchemaNode {
	displayText := schemaArgumentStyle.Render(name + " [block]")
	
	return &SchemaNode{
		id:          id,
		name:        name,
		displayText: displayText,
		path:        path,
		nodeType:    BlockNode,
		visible:     true,
		block:       block,
	}
}

// VisibleModel interface implementation
func (n *SchemaNode) IsVisible() bool {
	return n.visible
}

func (n *SchemaNode) Init() tea.Cmd {
	return nil
}

func (n *SchemaNode) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return n, nil
}

func (n *SchemaNode) View() string {
	return n.displayText
}

// Additional methods for our use case
func (n *SchemaNode) GetID() string {
	return n.id
}

func (n *SchemaNode) GetPath() []string {
	return n.path
}

func (n *SchemaNode) GetName() string {
	return n.name
}

func (n *SchemaNode) SetVisible(visible bool) {
	n.visible = visible
}

func (n *SchemaNode) GetAttribute() *tfjson.SchemaAttribute {
	return n.attribute
}

func (n *SchemaNode) GetBlock() *tfjson.SchemaBlock {
	return n.block
}

func (n *SchemaNode) IsAttribute() bool {
	return n.nodeType == AttributeNode
}

func (n *SchemaNode) IsBlock() bool {
	return n.nodeType == BlockNode
}
