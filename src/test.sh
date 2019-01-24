trap 'killAll' SIGINT

killAll(){
pkill -P $$
}

(go run main.go -port=4000) &
(go run main.go -port=4001)&
(go run main.go -port=4002) &

wait
