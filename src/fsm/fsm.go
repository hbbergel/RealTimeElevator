package fsm 

import "../elevio"
import "../types"
import "time"
import "fmt"
import "encoding/json"
import "io/ioutil"






func Fsm_run_elev(newOrder <-chan types.Button, floorReached <-chan int, orderDone chan types.Button, local_state chan types.ElevState) {

	

	e := types.ElevState{}
	{
		f := elevio.GetFloor()
		if f == -1 {
			e.Floor     = 0
			e.Direction = elevio.MD_Down
			e.State     = types.MOVING
			elevio.SetMotorDirection(elevio.MD_Down)
		} else {
			e.Floor     = f
			e.Direction = elevio.MD_Stop
			e.State     = types.IDLE
		}
	}
	
	
	var cabOrders [4]bool
	str, _ := ioutil.ReadFile("cabOrderBackup.json")		
	fmt.Printf("Reading from file")
	json.Unmarshal(str, &cabOrders)
	for f := 0; f < 4; f++ {
		if cabOrders[f] {
			e.Orders[f][2] = 1
			elevio.SetButtonLamp(2, f, true)
			e.State = types.INIT
		}
	}

	doorTime := time.NewTimer(3*time.Second)
	doorTime.Stop()

	
	fmt.Printf("Fsm has started")
	
	for{
		select{
		case newOrder := <- newOrder:

			e.Orders[newOrder.Floor][newOrder.Type] = 1
			local_state <- e
			//fmt.Printf("[ElevState]: Order:\n\t%+v\n", e.Orders)

			switch e.State {
			case types.IDLE:
				e.Direction = ChooseDirection(e)
				if (e.Direction == elevio.MD_Stop) && ShouldStop(e) {
					fmt.Printf("floor reached")
					e.State = types.DOOR_OPEN
					elevio.SetDoorOpenLamp(true)
					doorTime.Reset(3*time.Second)
					//time.Sleep(3*time.Second)
					
					local_state <- e

				} else {
					fmt.Println("Set Dir")
					elevio.SetMotorDirection(ChooseDirection(e))
					e.State = types.MOVING
					local_state <- e
				}
				
				
			case types.MOVING:
				e.Direction = ChooseDirection(e)
				local_state <- e

			case types.DOOR_OPEN:
				if e.Floor == newOrder.Floor {
					e.State = types.DOOR_OPEN
					doorTime.Reset(3*time.Second)
					local_state <- e

				}
			//case types.MOTOR_STOP:
				
			}
		
		case floorReached := <- floorReached:
			fmt.Println("Floor Arrival")
			elevio.SetFloorIndicator(floorReached)
			e.Floor = floorReached
			local_state <- e
			switch e.State {
			
			case types.MOVING:
				
				if ShouldStop(e) {
					elevio.SetMotorDirection(0)
					fmt.Println("Etasje!!!!")
					e.State = types.DOOR_OPEN
					elevio.SetDoorOpenLamp(true)
					e = ClearAtCurrentFloor(e, func(btn int){ orderDone <- types.Button{e.Floor, btn}})		
					doorTime.Reset(3*time.Second)
					//time.Sleep(3*time.Second)
					local_state <- e
				}
			case types.IDLE:
			case types.INIT:
				elevio.SetMotorDirection(0)
				e.State = types.IDLE
				e.Floor = floorReached
				e.Direction = elevio.MD_Stop
				local_state <- e
				fmt.Println("Initialisert")
			}
		case <- doorTime.C:
			
			switch e.State {
			case types.DOOR_OPEN:
				fmt.Printf("Door open")
				elevio.SetDoorOpenLamp(false)
				e.State = types.IDLE
				//fmt.Printf("Matrix,\n\t%+v\n", e.Orders)
				e = ClearAtCurrentFloor(e, func(btn int){ orderDone <- types.Button{e.Floor, btn}})		
				dir := ChooseDirection(e)
				elevio.SetMotorDirection(dir)
				e.Direction = dir
				if dir != elevio.MD_Stop {
					e.State = types.MOVING
				}
				local_state <- e
			}
		}
	}
}

func WriteCabOrdersToFile(localStateToFsm <-chan types.ElevState) {
	for{
		fmt.Printf("in for-loop")
		state := <- localStateToFsm
		var cabOrders [4]bool
		for f := 0; f < 4; f++ {
			if state.Orders[f][2] != 0 {
				cabOrders[f] = true
			}
		}
		str, _ := json.Marshal(cabOrders)
		fmt.Printf("writeing to file")
		ioutil.WriteFile("cabOrderBackup.json", str, 0666)
	}
}


//func Fsm_Init(local_id string, local_state chan<- types.ElevState, allStatesRx <-chan map[string]types.ElevState) {
//	elevio.SetMotorDirection(elevio.MD_Up)
//	
//	select {
//	case states, ok := <- allStatesRx:
//		
//		if ok {
//			
//			peerID := <- peerList
//			statesCpy := states[peerID[0]] 
//			e := states[local_id]
//			for floor := 0; floor <=3; floor++{
//				fmt.Printf("peer")
//				for btn := 0; btn <= 1; btn++{
//					if statesCpy.Orders[floor][btn] == 1 {
//						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, true)
//					} else if statesCpy.Orders[floor][btn] == 0 {
//						elevio.SetButtonLamp(elevio.ButtonType(btn), floor, false)
//					}
//					e.Orders[floor][btn] = 0
//				}
//			}
//			local_state <- e
//		}
//	default: 
//		
//		e := types.ElevState{
//			Floor: 0,
//			Direction: elevio.MD_Up,
//			State: types.INIT,
//			Orders: [types.N_FLOORS][types.N_BUTTONS]int {},
//		}
//		local_state <- e
//		
//
//	}
//	
//	/*
//	states := <- allStatesRx
//
//	if _, ok := states[local_id]; ok {
//		e := states[local_id]
//		local_state <- e
//		
//	} else {
//		e := types.ElevState{
//			Floor: 0,
//			Direction: elevio.MD_Up,
//			State: types.INIT,
//			Orders: [types.N_FLOORS][types.N_BUTTONS]int {},
//		}
//
//		local_state <- e
//	}
//
//	*/
//}



