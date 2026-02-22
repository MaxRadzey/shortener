-- wrk script: POST / with body = URL from profiles/urls.txt
-- If profiles/cookies.txt exists (generate with: go run cmd/gen_cookies/main.go), adds Cookie so all POSTs are under the same users.
-- Run from project root: wrk -t4 -c100 -d30s -s profiles/load_post.lua http://localhost:8080/

local urls = {}
local cookies = {}
local HOST = "localhost:8080"

function init(args)
  local f = io.open("profiles/urls.txt", "r")
  if not f then
    error("cannot open profiles/urls.txt (run wrk from project root)")
  end
  for line in f:lines() do
    line = line:gsub("^%s*(.-)%s*$", "%1")
    if #line > 0 then urls[#urls + 1] = line end
  end
  f:close()
  if #urls == 0 then
    error("profiles/urls.txt is empty")
  end

  f = io.open("profiles/cookies.txt", "r")
  if f then
    for line in f:lines() do
      line = line:gsub("^%s*(.-)%s*$", "%1")
      if #line > 0 then cookies[#cookies + 1] = line end
    end
    f:close()
  end
end

function request()
  local url = urls[math.random(#urls)]
  local body = url
  local header = "POST / HTTP/1.1\r\nHost: " .. HOST .. "\r\nContent-Type: text/plain\r\n"
  if #cookies > 0 then
    header = header .. "Cookie: user_id=" .. cookies[math.random(#cookies)] .. "\r\n"
  end
  return header .. "Content-Length: " .. #body .. "\r\n\r\n" .. body
end
