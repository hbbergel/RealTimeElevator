package fsm 

import "../elevio"
import "../types"
import "time"





func Fsm_run_elev(newOrder <-chan types.Button, floorReached <-chan int, orderDone chan<- types.Button, local_state chan<- types.ElevState) {

	doorTime := time.NewTimer(3*time.Second)

	e := types.ElevState{
		Floor: <-floorReached,
		Direction: elevio.MD_Stop,
		State: types.IDLE,
		Orders: [types.N_FLOORS][types.N_BUTTONS] int {},
	}
    
    local_state <- e



	for{
		select{
		case newOrder := <- newOrder:

			switch e.State {
			case types.IDLE:
				if e.Direction == elevio.MD_Stop {
					e.State = types.DOOR_OPEN
					elevio.SetDoorOpenLamp(true)
					time.Sleep(3*time.Second)
					elevio.SetDoorOpenLamp(false)
					e.State = types.IDLE
					local_state <- e

				} else {
					elevio.SetMotorDirection(ChooseDirection(e))
					e.State = types.MOVING
					local_state <- e
				}
				
				
			case types.MOVING:
			
			case types.DOOR_OPEN:
				if e.Floor == newOrder.Floor {
					elevio.SetDoorOpenLamp(true)
					time.Sleep(3*time.Second)
					elevio.SetDoorOpenLamp(false)
					e.State = types.IDLE
					local_state <- e

				}
			case types.MOTOR_STOP:
				
			}
		
		case <- floorReached:

			switch e.State {
			
			case types.MOVING:
				if ShouldStop(e) {
					ClearAtCurrentFloor(e, func(btn int){ orderDone <- types.Button{e.Floor, btn}})
					e.State = types.DOOR_OPEN
					elevio.SetMotorDirection(0)
					doorTime.Reset(3*time.Second)
					elevio.SetDoorOpenLamp(true)
					local_state <- e
				}				
			
			case types.INIT:
				elevio.SetMotorDirection(0)
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





