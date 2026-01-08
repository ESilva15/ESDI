// Package dom will be used to handle the multiple elements of an UI
package dom

import "github.com/rivo/tview"

// ErrElementAlreadyRegistered defines the error scenario of the user trying
// to register an element to the DOM with an already registered ID
type ErrElementAlreadyRegistered struct{}

func (e *ErrElementAlreadyRegistered) Error() string {
	return "an element with that ID has already been registered in the DOM"
}

type DOM struct {
	Root     *UINode
	elements map[string]*UINode
}

// NewDOM constructs a new DOM
func NewDOM() *DOM {
	return &DOM{
		Root:     nil,
		elements: make(map[string]*UINode),
	}
}

// Private
func (d *DOM) idIsRegisterd(ID string) bool {
	_, ok := d.elements[ID]
	return ok
}

// registerElem register a new element to the DOM
// should verify that the ID doesn't yet exist and thats it so far
func (d *DOM) registerElem(elem *UINode) error {
	if d.idIsRegisterd(elem.ID) {
		return &ErrElementAlreadyRegistered{}
	}

	d.elements[elem.ID] = elem
	return nil
}

// Public

// SetRoot defines the root element for the DOM
func (d *DOM) SetRoot(r *UINode) *DOM {
	d.Root = r
	return d
}

// NewUINode creates and registers a UI Node
func (d *DOM) NewUINode(ID string, parent, root tview.Primitive) (*UINode, error) {
	newNode := &UINode{
		ID:       ID,
		Self:     root,
		Children: make(map[string]*UINode),
	}

	err := d.registerElem(newNode)
	if err != nil {
		return nil, err
	}

	return newNode, nil
}

// GetRoot fetches the root UINode of the DOM
func (d *DOM) GetRoot(r *UINode) *UINode {
	return d.Root
}

// GetRootElem fetches the root primitive of the dom
func (d *DOM) GetRootElem() tview.Primitive {
	return d.Root.Self
}

// GetElemByID returns the primitive of a UINode given its ID
// returns nil if the Node isn't registered
func (d *DOM) GetElemByID(ID string) tview.Primitive {
	if !d.idIsRegisterd(ID) {
		return nil
	}

	return d.elements[ID].Self
}

// GetNodeByID returns a UINode given its ID
// returns nil if the Node isn't registered
func (d *DOM) GetNodeByID(ID string) *UINode {
	if !d.idIsRegisterd(ID) {
		return d.elements[ID]
	}

	return nil
}

// AppendElem adds a new UINode to a given UINode
func (d *DOM) AppendElem(root *UINode, elem *UINode) error {
	return root.AppendItem(elem)
}
