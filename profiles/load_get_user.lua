-- wrk script: GET /api/user/urls with Cookie from profiles/cookies.txt
-- Generate cookies first: go run cmd/gen_cookies/main.go
-- Then run POST load so those users have URLs; then: wrk -t4 -c100 -d30s -s profiles/load_get_user.lua http://localhost:8080/
local cookies = {}
local HOST = "localhost:8080"

function init(args)
  local f = io.open("profiles/cookies.txt", "r")
  if not f then
    error("cannot open profiles/cookies.txt (run from project root: go run cmd/gen_cookies/main.go)")
  end
  for line in f:lines() do
    line = line:gsub("^%s*(.-)%s*$", "%1")
    if #line > 0 then cookies[#cookies + 1] = line end
  end
  f:close()
  if #cookies == 0 then
    error("profiles/cookies.txt is empty")
  end
end

function request()
  local cookie = cookies[math.random(#cookies)]
  return "GET /api/user/urls HTTP/1.1\r\n" ..
         "Host: " .. HOST .. "\r\n" ..
         "Cookie: user_id=" .. cookie .. "\r\n\r\n"
end
