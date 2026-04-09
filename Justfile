set shell := ["fish", "-c"]
program_name := "gumu"

build:
  @go build -o {{program_name}}
  @go install .

run *args: build
  @{{program_name}} {{args}}

complete: build
  @./{{program_name}} completion -c fish | source

release:
  @go build -ldflags="-s -w" -o release/
  @ouch c -y --level 7 release/{{program_name}} release/{{program_name}}.tar.xz
