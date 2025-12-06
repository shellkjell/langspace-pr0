package runtime

import "time"

// timeNow is a variable that returns the current time.
// It's a variable so it can be mocked in tests.
var timeNow = time.Now
