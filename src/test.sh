trap 'killAll' SIGINT

killAll(){
pkill -P $$
}

export GOPATH=$(dirname $(pwd))

(go run main.go -port=4000)&
(go run main.go -port=4001)&
(go run main.go -port=4002)&
(go run main.go -port=4003)&
(go run main.go -port=4004)&
(go run main.go -port=4005)&
(go run main.go -port=4006)&
(go run main.go -port=4007)&
(go run main.go -port=4008)&
(go run main.go -port=4009)&

wait
