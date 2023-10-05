#!/bin/zsh
#
# Run script for Jolene

# get date and constant paths
DATE=`date +"%Y-%m-%d_%T"`
OAIPWD="/Users/rxn/dev/work/Jolene/backend/api/llama_fork/examples/server"
MPATH="~/.cache/lm-studio/models/TheBloke/Airoboros-L2-13B-2.1-GGUF/airoboros-l2-13b-2.1.Q5_K_S.gguf"
APIPWD="/Users/rxn/dev/work/Jolene/backend/api/"
TGPWD="/Users/rxn/dev/work/Jolene/backend/tgserver/"
MAINPWD="/Users/rxn/dev/work/Jolene/backend"
# Kill previous servers
killall server
killall main
echo "Killed previous processes"
echo "Made a new database"
(cd $TGPWD && rm database.db && touch database.db) 
# Start `llama_fork` servers
## Start LLaMa.cpp server
echo "Starting LLAMA server"
./api/llama_fork/server -m $MPATH -c 2048 -ngl 10 &> llama_server_${DATE}.log &

## Start OAI-like RestAPI server
echo "Starting OAI-like server"
(cd $OAIPWD && echo $PWD && $(./start.sh &> $MAINPWD/oai_server_${DATE}.log &))

# Start API server
echo "Starting API server"
(cd $APIPWD && go run main.go &| tee $MAINPWD/api_log_${DATE}.log &)

# Start Telegram server
echo "Starting Telegram server"
(cd $TGPWD && go run main.go &| tee $MAINPWD/tgserver_log_${DATE}.log)
