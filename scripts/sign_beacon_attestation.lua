-- Attempt to approve a request to sign a beacon attestation.
--
-- Contents of request are as follows:
--   - ip: the IP address of the requesting server (string)
--   - client: the name of the client presenting the request, as defined by its certificate (string)
--   - timestamp: the Unix timestamp of the request (number)
--   - account: the name of the account presenting the request (string)
--   - pubKey: the public key of the account presenting the request (string)
--   - domain: the domain of the request (string)
--   - slot: the slot of the request (number)
--   - committeeIndex: the committe index of the account for the attestation (number)
--   - sourceEpoch: the source epoch being attested (number)
--   - sourceRoot: the source state root being attested (number)
--   - targetEpoch: the target epoch being attested (number)
--   - targetRoot: the target state root being attested (number)
--  storage is storage that persists between calls, specific to this request type and account
--  messages is a list of messages that will be printed in the logs at the conclusion of the script
function approve(request, storage, messages)
  return "Approved"
end
