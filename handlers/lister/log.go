package lister

import zerologger "github.com/rs/zerolog/log"

var log = zerologger.With().Str("module", "lister").Logger()
