-- wrk script: GET /:short_path from profiles/short_paths.txt
-- Run from project root: wrk -t4 -c100 -d30s -s profiles/load_get.lua http://localhost:8080/

local paths = {}
local HOST = "localhost:8080"

function init(args)
  local f = io.open("profiles/short_paths.txt", "r")
  if not f then
    error("cannot open profiles/short_paths.txt (run wrk from project root, generate with: go run cmd/gen_short_paths/main.go)")
  end
  for line in f:lines() do
    line = line:gsub("^%s*(.-)%s*$", "%1")
    if #line > 0 then
      paths[#paths + 1] = line
    end
  end
  f:close()
  if #paths == 0 then
    error("profiles/short_paths.txt is empty")
  end
end

function request()
  local path = paths[math.random(#paths)]
  return "GET /" .. path .. " HTTP/1.1\r\n" ..
         "Host: " .. HOST .. "\r\n\r\n"
end
