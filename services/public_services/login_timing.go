package public_services

import "github.com/pierceperado/smpc/utils"

// One message for an unknown employee id and for a wrong password. They used to
// differ ("Invalid user employee id" / "Invalid user password"), which told anyone
// trying ids which ones exist - and ids follow a predictable DEPARTMENT-POSITION-number
// pattern. No client reads either message.
const invalidLoginMessage = "Invalid employee id or password"

// A bcrypt hash of a password nobody uses. An unknown id is compared against it, so
// the request costs the same bcrypt work as a real account and the response time does
// not give the answer away instead.
var dummyPasswordHash, _ = utils.GenerateUserPassword("lightspeed-login-timing-placeholder")
