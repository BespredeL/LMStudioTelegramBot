# Telegram Bot + LM Studio

[![Readme EN](https://img.shields.io/badge/README-EN-blue.svg)](https://github.com/bespredel/LMStudioTelegramBot/blob/master/README.md)
[![Readme RU](https://img.shields.io/badge/README-RU-blue.svg)](https://github.com/bespredel/LMStudioTelegramBot/blob/master/README_RU.md)
[![GitHub license](https://img.shields.io/badge/license-MIT-458a7b.svg)](https://github.com/bespredel/LMStudioTelegramBot/blob/master/LICENSE)

This repository contains a Go application for running a Telegram bot integrated with LM Studio. The bot can be configured to use either polling or webhook methods to interact with users. It features a graphical user interface (GUI) built with the [Fyne framework](https://fyne.io/), allowing users to configure bot settings, select models, and manage users.

## Features

- **Bot control**: Launch and stop the Telegram bot from the interface.
- **Configuration settings**: Easily configure the bot's API address, token, update method (polling/webhook), webhook details, and more.
- **Model management**: Select a model from the available options, and refresh the model list.
- **User management**: View and update the allowed status of users in the system.
- **Logs**: View bot logs in real-time.

## Requirements

- Go (1.19+)
- Fyne (GUI framework)
- Telegram bot token (from [BotFather](https://core.telegram.org/bots#botfather))

## Installation

1. Clone this repository:
    ```sh
    git clone https://github.com/yourusername/telegram-bot-lmstudio.git
    cd telegram-bot-lmstudio
    ```

2. Install the required Go dependencies:
    ```sh
    go get fyne.io/fyne/v2
    ```

3. Build and run the application:
    ```sh
    go run main.go
    ```

## Configuration

Upon launching the application, the user is presented with several configuration options:

- **API Address**: The address of the LM Studio API (e.g., `http://localhost:1234`).
- **Max Tokens**: The token limit for the model.
- **Timeout**: The polling timeout in seconds.
- **Bot Token**: The Telegram bot token obtained from BotFather.
- **Update Method**: Choose between "polling" or "webhook" for receiving updates.
- **Webhook Domain/Port**: Details required for setting up a webhook (only for webhook method).
- **System Role**: The system role used in the LM Studio configuration.
- **LM Studio Mode**: Select between "stream" or "full" modes for interacting with LM Studio.
- **Language**: Choose the language for the bot (e.g., English or Russian).

## Bot Control

You can start and stop the Telegram bot from the **Bot** tab. The bot interacts with Telegram users based on the selected model and configuration. The bot logs all interactions and displays them in real-time on the GUI.

- **Polling**: The bot regularly checks for new messages.
- **Webhook**: The bot listens for incoming requests on the specified webhook domain and port.

## User Management

The **Users** tab displays a list of users, their ID, username, and whether they are allowed to interact with the bot. You can enable or disable user access by updating the "Allowed" checkbox.

## Logs

The **Bot** tab also displays logs, showing the requests received by the bot. The logs are refreshed every 2 seconds to provide real-time updates.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.