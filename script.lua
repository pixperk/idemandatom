-- script.lua
local key = KEYS[1]
local pending_status = ARGV[1] -- Value: "PENDING"
local ttl = ARGV[2]            -- Expiration: e.g., 60s

-- 1. Check if key exists
local value = redis.call("GET", key)

-- 2. If it exists, return the value 
-- (This could be "PENDING" or the final JSON response)
if value then
    return value 
end

-- 3. If not, lock it with "PENDING" status and a TTL
-- The TTL is crucial: If our server crashes mid-process, 
-- this key must eventually expire so the user can retry.
redis.call("SET", key, pending_status, "EX", ttl)
return nil