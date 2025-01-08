package extensions

import (
	"fmt"

	"github.com/chriscow/minds"
	lua "github.com/yuin/gopher-lua"
)

// LuaCondition implements SwitchCondition by evaluating a Lua script. The script is
// executed in a fresh Lua state for each evaluation and must return a boolean value.
//
// The script has access to the following global variables:
//   - metadata: A string representation of the thread context's metadata map
//   - last_message: The content of the last message in the thread (if available)
//
// Example usage:
//
//	condition := LuaCondition{
//	    Script: `
//	        -- Check if metadata contains "type=question"
//	        return string.find(metadata, "type=question") ~= nil
//	    `,
//	}
//	result, err := condition.Evaluate(threadContext)
type LuaCondition struct {
	// Script contains the Lua code to be executed.
	// The script must return a boolean value as its final result.
	Script string
}

// Evaluate executes the Lua script and returns its boolean result. It creates a new
// Lua state, populates it with thread context data, executes the script, and
// interprets the result.
//
// The function provides the following global variables to the Lua environment:
//   - metadata: String representation of tc.Metadata()
//   - last_message: Content of the last message in tc.Messages() (if any)
//
// Returns an error if:
//   - The script execution fails
//   - The script returns a non-boolean value
func (l LuaCondition) Evaluate(tc minds.ThreadContext) (bool, error) {
	// Create a new Lua state for this evaluation
	L := lua.NewState()
	defer L.Close()

	// Expose thread context to Lua environment
	// Convert metadata to string since Lua has limited type support
	L.SetGlobal("metadata", lua.LString(fmt.Sprintf("%v", tc.Metadata())))

	// Make the last message content available if the thread has messages
	if len(tc.Messages()) > 0 {
		L.SetGlobal("last_message", lua.LString(tc.Messages()[len(tc.Messages())-1].Content))
	}

	// Execute the Lua script
	if err := L.DoString(l.Script); err != nil {
		return false, fmt.Errorf("error executing Lua script: %w", err)
	}

	// Get the result from the top of the Lua stack
	result := L.Get(-1) // Get last value
	L.Pop(1)            // Remove it from stack

	// Ensure the result is a boolean and convert it to Go bool
	switch v := result.(type) {
	case lua.LBool:
		return bool(v), nil
	default:
		return false, fmt.Errorf("lua script must return a boolean")
	}
}
