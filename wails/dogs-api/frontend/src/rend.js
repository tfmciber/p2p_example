import { AddRendezvous } from "../wailsjs/go/main/P2Papp.js";
import { CancelRendezvous } from "../wailsjs/go/main/P2Papp.js";
import { getModal } from "./Modals.svelte";
export async function addRend() {

    let rend = document.getElementById("rend").value;

   
    AddRendezvous(rend).then();
    document.forms["rendform"].reset();

    getModal('rendmodal').close();

  };
  export function reload(chat) {

    AddRendezvous(chat).then();
  }
  export async function cancelRendezvous() {
    let rend = document.getElementById("rend").value;
    CancelRendezvous(rend).then();
    }
    export function cancelRendezvousstr(rend) {
      CancelRendezvous(rend).then();
      }




