-- wrk script: POST / with body = URL from profiles/urls.txt
-- Run from project root: wrk -t4 -c100 -d30s -s profiles/load_post.lua http://localhost:8080/

local urls = {}
local HOST = "localhost:8080"

function init(args)
  local f = io.open("profiles/urls.txt", "r")
  if not f then
    error("cannot open profiles/urls.txt (run wrk from project root)")
  end
  for line in f:lines() do
    line = line:gsub("^%s*(.-)%s*$", "%1")
    if #line > 0 then
      urls[#urls + 1] = line
    end
  end
  f:close()
  if #urls == 0 then
    error("profiles/urls.txt is empty")
  end
end

function request()
  local url = urls[math.random(#urls)]
  local body = url
  return "POST / HTTP/1.1\r\n" ..
         "Host: " .. HOST .. "\r\n" ..
         "Content-Type: text/plain\r\n" ..
         "Content-Length: " .. #body .. "\r\n\r\n" ..
         body
end
