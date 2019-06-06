package errors

import (
    "fmt"
)
var (
    OP_SUCCESS = fmt.Errorf("Operation is success")
    CMD_NOT_FOUND = fmt.Errorf("Command not found")
    INVALID_INPUT = fmt.Errorf("Invalid Input")
    INVALID_OP = fmt.Errorf("Invalid operation request")
    DATA_NOT_UNIQUE_ERROR = fmt.Errorf("The entry is not unique in the App")
    DATA_PRESENT_IN_SYSTEM = fmt.Errorf(`The entry already present in App`)
    DATA_NOT_FOUND = fmt.Errorf("The entry not found in the Application")
)
