require 'json'

body = JSON.parse(STDIN.read)
STDOUT.write JSON.generate(body)

