all:
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver/testclient
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver/simsdk
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver/simapp

pushserver:
	go install github.com/zhengzhiren/pushserver

testclient:
	go install github.com/zhengzhiren/pushserver/testclient

simsdk:
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver/simsdk

simapp:
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver/simapp

clean:
	go clean github.com/zhengzhiren/pushserver
	go clean github.com/zhengzhiren/pushserver/testclient
	go clean github.com/zhengzhiren/pushserver/simsdk
	go clean github.com/zhengzhiren/pushserver/simapp

test:
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver/simsdk
	go install -gcflags "-N -l" -ldflags "-s" github.com/zhengzhiren/pushserver/simapp
	mkdir -p temp
	cp $(GOPATH)/bin/simsdk temp
	cp $(GOPATH)/bin/simapp temp
	cp scripts/load.sh temp
	cp scripts/log_proc.sh temp
	cp scripts/sys_tune.sh temp
	cd temp; tar -czf test.tgz load.sh log_proc.sh simapp simsdk sys_tune.sh; mv test.tgz ../
