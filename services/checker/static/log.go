package static

import zerologger "github.com/rs/zerolog/log"

var log = zerologger.With().Str("module", "staticchecker").Logger()
