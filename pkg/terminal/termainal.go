package terminal

import (
	"bufio"
	"fmt"
	sdk "liangyuanguo/aw/mr/pkg/client"
	api "liangyuanguo/aw/mr/pkg/model"
	"os"
	"strconv"
	"strings"
)

func GetTerminalInput(e *sdk.EndPoint, roomId int64) {
	scanner := bufio.NewScanner(os.Stdin)

	toGroup := 1
	var toUser = []int64{1}

	//e.SendQ <- &api.Message{
	//	ToGroup: 2,
	//	Msg: "join-room 1",
	//}

	//go func() {
	//	for {
	//		select {
	//		case msg := <-e.RevQ:
	//			jsonString, _ := json.Marshal(msg)
	//			if msg.From == api.SuperUserId {
	//				fmt.Printf("%s: %s\n", api.ParseCmd(msg.Msg), jsonString)
	//				continue
	//			}
	//			fmt.Printf("rev Json: %s\n", jsonString)
	//		}
	//	}
	//}()

	for {
		fmt.Print("请输入内容: ")
		scanner.Scan()
		input := scanner.Text()

		cmd := api.Parse(input)
		if len(cmd.Args) <= 0 {
			continue
		}

		if cmd.Args[0] == "quit" {
			fmt.Println("退出循环")
			os.Exit(0)
		} else if cmd.Args[0] == "help" {
			fmt.Println(`commanders:
				1. quit
				2. set [-to=1,2,3,4]]
				3. rsys 
			`)
			continue
		} else if cmd.Args[0] == "set" {

			toGroup = cmd.GetOptInt("g", toGroup)
			toUser = []int64{}
			toUser2 := cmd.GetOptIntArr("to")

			for _, uid := range toUser2 {
				toUser = append(toUser, int64(uid))
			}

			var users []string
			for _, uid := range toUser {
				users = append(users, strconv.Itoa(int(uid)))
			}
			fmt.Println(fmt.Sprintf("roomId=%d from=%d to=%s group=%d", roomId, e.User.Uid, strings.Join(users, ","), toGroup))
			continue
		} else if cmd.Args[0] == "rsys" {
			cmd.Args = cmd.Args[1:]
			msg, err := api.ParseCmd(cmd)
			if err != nil {
				fmt.Println(err)
				continue
			}
			e.SendQ <- msg
		} else {
			e.SendQ <- &api.Message{
				Room: roomId,
				From: e.User.Uid,
				Msg:  input,
				To:   toUser,
			}
		}

	}
}
