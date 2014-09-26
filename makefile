all:
	go install github.com/zhengzhiren/pushserver
	go install github.com/zhengzhiren/pushserver/testclient
	go install github.com/zhengzhiren/pushserver/simsdk
	go install github.com/zhengzhiren/pushserver/simapp

pushserver:
	go install github.com/zhengzhiren/pushserver

testclient:
	go install github.com/zhengzhiren/pushserver/testclient

clean:
	go clean github.com/zhengzhiren/pushserver
	go clean github.com/zhengzhiren/pushserver/testclient
