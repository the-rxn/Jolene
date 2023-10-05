#!/bin/zsh
#
# Run script for Jolene

# get date and constant paths
DATE=`date +"%Y-%m-%d_%T"`
OAIPWD="/Users/rxn/dev/work/Jolene/backend/api/llama_fork/examples/server"
MPATH="/Users/rxn/.cache/lm-studio/models/TheBloke/Airoboros-L2-13B-2.1-GGUF/airoboros-l2-13b-2.1.Q5_K_S.gguf"
APIPWD="/Users/rxn/dev/work/Jolene/backend/api/"
TGPWD="/Users/rxn/dev/work/Jolene/backend/tgserver/"
MAINPWD="/Users/rxn/dev/work/Jolene/backend"
LOGSPWD="${MAINPWD}/logs"
echo "Logs located in: ${LOGSPWD}"
# Kill previous servers
killall server
killall main
echo "Killed previous processes"
echo "Made a new database"
(cd $TGPWD && rm database.db && touch database.db) 
# Start `llama_fork` servers
## Start LLaMa.cpp server
echo "Starting LLAMA server"
CMD="./api/llama_fork/server -m $MPATH -c 2048 -ngl 10 &| tee $LOGSPWD/llama_server_$DATE.log &"
echo $CMD
eval $CMD


## Start OAI-like RestAPI server
echo "Starting OAI-like server"
CMD="(cd $OAIPWD && echo $PWD && $(./start.sh &| tee $LOGSPWD/oai_server_$DATE.log &))"
echo $CMD
eval $CMD



# Start API server
echo "Starting API server"
CMD="(cd $APIPWD && go run main.go &| tee $LOGSPWD/api_log_$DATE.log &)"
echo $CMD
eval $CMD

# Start Telegram server
echo "Starting Telegram server"
CMD="(cd $TGPWD && go run main.go &| tee $LOGSPWD/tgserver_log_$DATE.log)"
echo $CMD
eval $CMD
