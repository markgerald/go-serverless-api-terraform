.PHONY: get tidy build synth diff deploy destroy clean

get:
	cdktf get

tidy:
	go mod tidy

build:
	go build ./...

synth:
	cdktf synth

diff:
	cdktf diff

deploy:
	cdktf deploy

destroy:
	cdktf destroy

clean:
	rm -rf cdktf.out .terraform .terraform.lock.hcl
