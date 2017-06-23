package mem

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCategory(t *testing.T) {
	ctx := NewContext()
	list := ctx.VideoCategory.FindAll()
	b, _ := json.Marshal(list)
	fmt.Printf("%v", string(b))
}
