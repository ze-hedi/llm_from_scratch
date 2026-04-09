# Extensions

This directory contains extensions for the chatbot-tui application. Each extension is a self-contained module that adds new functionality to the main application.

## Browsing Extensions

You can browse and launch available extensions using:

```bash
./chatbot-tui extensions
```

This will display an interactive list of all available extensions. Use arrow keys to navigate and press Enter to launch.

## Available Extensions

### Tamagotchi

A virtual pet game where you can care for and interact with a digital companion.

**Commands:**
- `./chatbot-tui tamagotchi` - Start playing with your pet (or choose a new one if you don't have one)
- `./chatbot-tui tamagotchi choose` - Choose a new pet
- `./chatbot-tui tamagotchi reset` - Delete your current pet and start fresh

**How to Play:**
1. Run `./chatbot-tui tamagotchi` to start
2. If this is your first time, you'll be prompted to choose a pet (Mochi the cat, Lucy the turtle, or Ottopus the octopus)
3. Use commands to interact with your pet:
   - `feed` - Give your pet food
   - `play` - Play with your pet
   - `heal` - Heal your pet when it's sick
   - `status` - Check your pet's current stats
   - `quit` - Exit the game

Your pet's stats degrade over time, so make sure to check on them regularly!

## Creating New Extensions

To create a new extension:

1. Create a new directory under `extensions/`
2. Implement your extension's logic
3. Create a new command file in `cmd/` that integrates your extension
4. Add your extension to `extensions.json` in the project root:
   ```json
   {
     "id": "your-extension-id",
     "name": "Extension Name",
     "description": "Brief description of what it does",
     "command": "command-name",
     "icon": "🎮"
   }
   ```
5. Update the `cmd/extensions.go` file to handle your new command
6. Update this README with information about your extension
