package fsm 

import "../elevio"
import "../types"
import "time"
import "fmt"





func Fsm_run_elev(newOrder <-chan types.Button, floorReached <-chan int, orderDone chan<- types.Button, local_state chan<- types.ElevState) {
	
	
	elevio.SetMotorDirection(elevio.MD_Up)
	elevio.SetStopLamp(false)
	e := types.ElevState{
		Floor: 0,
		Direction: elevio.MD_Up,
		State: types.INIT,
		Orders: [types.N_FLOORS][types.N_BUTTONS]int {},
	}
	
	doorTime := time.NewTimer(3*time.Second)


	local_state <- e
	fmt.Printf("Fsm has started")
	
	for{
		select{
		case newOrder := <- newOrder:

			e.Orders[newOrder.Floor][newOrder.Type] = 1

			switch e.State {
			case types.IDLE:
				if e.Direction == elevio.MD_Stop {
					e.State = types.DOOR_OPEN
					elevio.SetDoorOpenLamp(true)
					doorTime.Reset(3*time.Second)
					local_state <- e

				} else {
					elevio.SetMotorDirection(ChooseDirection(e))
					e.State = types.MOVING
					local_state <- e
				}
				
				
			case types.MOVING:
			
			case types.DOOR_OPEN:
				if e.Floor == newOrder.Floor {
					e.State = types.DOOR_OPEN
					doorTime.Reset(3*time.Second)
					local_state <- e

				}
			case types.MOTOR_STOP:
				
			}
		
		case floorReached := <- floorReached:
			fmt.Println("Etasje!!!!")
			elevio.SetFloorIndicator(floorReached)
			switch e.State {
			
			case types.MOVING:
				if ShouldStop(e) {
					ClearAtCurrentFloor(e)  //, func(btn int){ orderDone <- types.Button{e.Floor, btn}})
					e.State = types.DOOR_OPEN
					elevio.SetMotorDirection(0)
					doorTime.Reset(3*time.Second)
					elevio.SetDoorOpenLamp(true)
					local_state <- e
				}				
			
			case types.INIT:
				elevio.SetMotorDirection(0)
				e.State = types.IDLE
				local_state <- e
			}
		case <- doorTime.C:
			
			switch e.State {
			case types.DOOR_OPEN:
				elevio.SetDoorOpenLamp(false)
				dir := ChooseDirection(e)
				elevio.SetMotorDirection(dir)
				local_state <- e
			}
		
		}
	}
	

}





