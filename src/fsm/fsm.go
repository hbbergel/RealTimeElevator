package fsm 

import "../elevio"
import "../types"
import "time"
import "fmt"





func Fsm_run_elev(newOrder <-chan types.Button, floorReached <-chan int, orderDone chan types.Button, local_state chan types.ElevState) {
	
	e := <- local_state
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
					doorTime.Reset(3*time.Second)
					//time.Sleep(3*time.Second)
					local_state <- e
				}
			
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

func Fsm_Init(local_id string, local_state chan<- types.ElevState, allStates <-chan map[string]ElevState) {
	elevio.SetMotorDirection(elevio.MD_Up)

	if val, ok := allStates[local_id]; ok {
		e := allStates[local_id]
		local_state <- e
		
	} else {
		e := types.ElevState{
			Floor: 0,
			Direction: elevio.MD_Up,
			State: types.INIT,
			Orders: [types.N_FLOORS][types.N_BUTTONS]int {},
		}
		local_state <- e
	}

	
}



