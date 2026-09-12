package blocking

import "errors"

// ErrCannotBlockSelf: no tiene sentido bloquearte a ti mismo.
var ErrCannotBlockSelf = errors.New("blocking: no puedes bloquearte a ti mismo")
