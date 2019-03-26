package elevstates

import "../types"
import "../bcast"
import "time"
import "fmt"
	


type T struct {
	State types.ElevState
	ID string
}


func ElevStates(local_id string, local_state <-chan types.ElevState, all_states chan<- map[string]types.ElevState){

	states := make(map[string]types.ElevState)
	
	ticker := time.NewTicker(time.Millisecond)
	netSend := make(chan T)
	netRecv := make(chan T)

    go bcast.Transmitter(15001, netSend)
	go bcast.Receiver(15001, netRecv)



    for{
        select{
		case a := <- local_state:
			if a != states[local_id] {
				fmt.Printf("[ElevState]: Recieved new local state:\n\t%+v\n", a)
				states[local_id] = a
				all_states <- states
			}

			netSend <- T{a, local_id}

		case a := <- netRecv:
			if a.ID != local_id{
				remoteState, ok := states[a.ID]
				if !ok || remoteState != a.State {
					fmt.Printf("[ElevStates]: Received new remote state:\n\t%+v\n", a)
                    states[a.ID] = a.State
                    all_states <- states
				}
			}
		case <- ticker.C:
			if localState, ok := states[local_id]; ok{
				netSend <- T{localState, local_id}
			} 

		}
		
    }
}