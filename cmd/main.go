package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"log"
)

func main() {

	client := telegram.NewClient(apiID, apiHash, telegram.Options{})

	err := client.Run(context.Background(), func(ctx context.Context) error {

		authStatus, err := client.Auth().Status(ctx)

		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}

		if !authStatus.Authorized {
			// https://github.com/gotd/td/tree/master/examples
			return fmt.Errorf("not authorized: you need to sign in")
		}

		wrappedTgClient := tg.NewClient(client)

		log.Println("Client is authorized and ready!")
		return nil
	})

	if err != nil {
		log.Fatalf("Error running client: %v", err)
	}
}

// Создаём клиент
// Запускаем клиент, оборачивая всё в client.Run
// Проверка/выполнение аутентификации
// Получаем список чатов или конкретный чат
// (Например, настроить в конфиге, какие чаты интересуют)
// Загружаем сообщения и скачиваем медиа
// сортировочка
