package elevstates

import "../types"
import "../bcast"
import "time"
	


type T struct {
	State types.ElevState
	ID string
}


func ElevStates(local_id string, local_state <-chan types.ElevState, all_states chan<- map[string]types.ElevState ){

	var states map[string]types.ElevState
	
	ticker := time.NewTicker(100*time.Millisecond)
	var l_state T 
	netSend := make(chan T)
	netRecv := make(chan T)

    go bcast.Transmitter(15001, netSend)
	go bcast.Receiver(15001, netRecv)

    for{
        select{
		case a := <- local_state:
			states[local_id] = a
			l_state = T{a, local_id}
			netSend <- l_state


		case a := <- netRecv:
			if a.ID != local_id{
					states[a.ID] = a.State
			}
		case <- ticker.C:

			netSend <- l_state
				// do lights here: hall for all elevs, cab for us 
				
			
		case all_states <- states:

	

		}
		
    }
}