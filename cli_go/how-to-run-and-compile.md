# How to compile and run

## Compile

```bash
go build -o otto-chat ./cmd/otto-chat/
go build -o otto-tmux ./cmd/otto-tmux/
```

## Run

### Chat (TUI chatbot + extensions)

```bash
./otto-chat chat          # start chat UI
./otto-chat settings      # configure model settings
./otto-chat extensions    # browse extensions
./otto-chat tamagotchi    # play tamagotchi
./otto-chat dino          # play dino runner
```

### Tmux (session manager)

```bash
./otto-tmux three                  # interactive pane setup
./otto-tmux dev                    # editor + 3 shells layout
./otto-tmux custom --panes 4      # N blank panes
./otto-tmux attach -s mysession   # attach to existing session
./otto-tmux kill -s mysession     # kill a session
```
