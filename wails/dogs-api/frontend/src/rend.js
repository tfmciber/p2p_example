import { AddRendezvous } from "../wailsjs/go/main/P2Papp.js";
import { CancelRendezvous } from "../wailsjs/go/main/P2Papp.js";
export async function addRend() {
    const loader = document.querySelector(".loader");
    const submitBtn = document.getElementById("submit-btn");
    const cancelBtn = document.getElementById("cancel-btn");
    let rend = document.getElementById("rend").value;
    

    //show loader
    loader.style.display = "block";
    cancelBtn.style.display = "block";
    submitBtn.style.display = "none";
   
    await AddRendezvous(rend).then();
    document.forms["rendform"].reset();
    
    loader.style.display = "none";
    cancelBtn.style.display = "none";
    submitBtn.style.display = "";
  };
  export async function cancelRendezvous() {
    let rend = document.getElementById("rend").value;
    CancelRendezvous(rend).then();
    }




