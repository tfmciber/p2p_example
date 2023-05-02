
import progressicon from "./assets/images/progress.png";
import check from "./assets/images/check.png";
import wrong from "./assets/images/wrong.png";
import queue from "./assets/images/queue.png";
import { OpenFileExplorer } from "../wailsjs/go/main/P2Papp.js";
import fileIcon from "./assets/images/folder.png";

export function getColorForUserId(userId) {
    // Convert the user ID to a number
    const num = parseInt(userId, 36);

    // Choose a large prime number as the modulus
    const prime = 65537;

    // Generate random RGB values between 0 and 255
    const r = Math.floor(Math.abs(Math.sin(num) * prime) % 256);
    const g = Math.floor(Math.abs(Math.cos(num) * prime) % 256);
    const b = Math.floor(Math.abs(Math.tan(num) * prime) % 256);

    // Return the random color in RGB format
    return `rgb(${r}, ${g}, ${b})`;
  };

  export function checkPasswordMatch(passwordid, confirmPasswordid, passwordMatchErrorid,loginformid) {
    const passwordInput = document.querySelector(passwordid);
    const confirmPasswordInput = document.querySelector(confirmPasswordid);
    const passwordMatchError = document.querySelector(passwordMatchErrorid);
    const loginform = document.getElementById(loginformid);

    var submitButton = loginform.querySelector('button[type="submit"]');


    if (passwordInput.value !== confirmPasswordInput.value) {
      submitButton.disabled = true;
      passwordMatchError.style.display = "block";
    } else {
      submitButton.disabled = false;
      passwordMatchError.style.display = "none";
    }
  };
  export function createCommand(cmd) {
    let terminal = document.getElementById("terminal-box");
    let command = document.createElement("div");
    command.className = "command";
    command.innerHTML = cmd;
    terminal.scrollTop = terminal.scrollHeight + 20;
    terminal.appendChild(command);
  };

 export async function updateProgress(rend, peer, progress, fileName) {
    let progressbar = document.getElementById("progress" + rend + peer+fileName);
    let progressbuttons = document.getElementById("progressbuttons" + rend + peer+fileName);
    //if there are several progress bars for the same rendezvous, the first one will be the one that is updated
    
    
        progressbar.style.width = progress + "%";
       
        if (progress == 0){
        progressbuttons.innerHTML= `<img class="message-ok" src=${progressicon} alt="ok" />`;
      }
        if (progress == 100) {
        // progressbar.style.backgroundColor = "green";
         progressbuttons.innerHTML= `<img class="message-ok" src=${check} alt="ok" />`;
         progressbar.id = "done";
         progressbuttons.id = "done";
        
        }
        if (progress == -1)
        {
                   
          progressbar.id = "error";
          progressbuttons.id = "error";
          progressbuttons.innerHTML= `<img class="message-ok" src=${wrong} alt="ok" />`;
        }
  };


  export async function SetUsers(Users2,current_red, directmessages) {
    if (current_red == "settings"){
      current_red = "";
    }
    let usersAux = Users2[current_red];
    if (directmessages.includes(current_red)) {
      usersAux = Users2[""].filter((user) => user.user == current_red);
    }

    //set users in html periodically in right side
    let userscurrent = document.getElementById("users");
    //get offline users container inside the users container
    let offline = userscurrent.getElementsByClassName("offlineusers")[0];
    let online = userscurrent.getElementsByClassName("onlineusers")[0];
    offline.innerHTML = "";
    online.innerHTML = "";
    let onlinenum = 0;
    let offlinenum = 0;
    if (usersAux != null) {
      onlinenum = usersAux.filter((user) => user.online).length;
      offlinenum = usersAux.length - onlinenum;
    }
    if (onlinenum > 0) {
      online.innerHTML =
        "<h1 class='onlineheader'>Online - " + onlinenum + " </h1>";
    }
    if (offlinenum > 0) {
      offline.innerHTML =
        "<h1 class='offlineheader'>Offline - " + offlinenum + "</h1>";
    }
    if (usersAux != null) {
      usersAux.forEach((user) => {
        if (user.online) {
          //add div with onlick event listener
          online.innerHTML += `<div class="useronline" id="${user.user}">${user.user}</div>`;
        } else {
          offline.innerHTML += `<div class="useroffline" id="${user.user}">${user.user}</div>`;
        }
      });
    }

    //add event listener for every user, online and offline

    let allusers = userscurrent.getElementsByClassName("useronline");
    let allusers2 = userscurrent.getElementsByClassName("useroffline");
    let users = [...allusers, ...allusers2];

    for (let i = 0; i < users.length; i++) {
      users[i].addEventListener("click", function (event) {
        showpopup(users[i].id, event);
      });
    }
  };

  function showpopup(text, event) {
    var modal = document.getElementById("popup");
    var buttonRect = event.target.getBoundingClientRect();
    var buttonTop = buttonRect.top + window.scrollY;
    var poputname = modal.querySelector("#popupname");
    poputname.innerText = text;

    modal.style.top = buttonTop + "px";
    modal.style.display = "block";
    modal.style.right = "300px";
  }

    //creates a new message div and returns it
    export function createMessage(chat, message, sender, time, files, ok) {
    
     
      let chatbox = document.getElementById("chat-box" + chat);
      
      let newmessage = document.createElement("div");
      if (sender == "me") {
        newmessage.className = "messagesent";
        //get time and date dd/mm/yyyy hh:mm
        
      } else {
        newmessage.className = "messagereceived";
      }
      //if files are not null, add them to the message
  
      let container = document.createElement("div");
      container.className = "fileIconmessageSCONTAINER";
      if (files != null) {
  
        //loop over files
  
        for (let i = 0; i < files.length; i++) {
  
  
          let text = document.createElement("div");
          text.innerText = files[i].filename;
          text.className = "textmessage";
  
          let button = document.createElement("img");
          // Add a click event listener to the button
          button.addEventListener("click", async () => {
            //get path attribute
            let path = button.getAttribute("path");
            await OpenFileExplorer(path).then();
          });
          //add and image inside the button
          let name = document.createElement("div");
          name.innerText = files[i].filename;
          button.src = fileIcon;
          button.className = "fileIconmessage";
          let file = files[i];
          let path = file.path;
          //add path attribute to button
          button.setAttribute("path", path);
          //create progress bar
          let progress = document.createElement("div");
          progress.className = "progress";
          let progressbar = document.createElement("div");
          progressbar.className = "progressbar";
          progressbar.id = "progress" + chat + sender + files[i].filename;
          
          let progressbuttons = document.createElement("div");
          progressbuttons.id = "progressbuttons" + chat + sender + files[i].filename;
          progressbuttons.className = "progress-buttons";
          progressbuttons.innerHTML= `<img class="message-ok" src=${queue} alt="ok" />`;
          progress.appendChild(progressbar);
          container.appendChild(button);
          container.appendChild(text);
          container.appendChild(progress);
          container.appendChild(progressbuttons);
        }
      }
  if (message != ""){
      newmessage.innerHTML = `<div class="message-header">
      <div class="message-sender">${sender}</div>
      <div class="message-time">${time}</div>
    </div>
      <div class="message-text">${message}</div>
      `;
      if (sender == "me" ) {
        if (ok == true) {
          newmessage.innerHTML += `<img class="message-ok" src=${check} alt="ok" />`;
        } else {
          newmessage.innerHTML += `<img class="message-ok" src=${wrong} alt="ok" />`;
        }
      }
  }
      if (files != null) {
        newmessage.appendChild(container);
      }
  
      chatbox.appendChild(newmessage);
      //scroll to bottom
      chatbox.scrollTop = chatbox.scrollHeight;
    }

  
   export function auxchangechat(last_rend,current_redaux,Users,directmessages) {


      var currentchatdivid;
      var lastchatdivid;
      if (current_redaux == "settings")
       {
        currentchatdivid="settings";
      } else if (current_redaux != "") {
        currentchatdivid="chat" + current_redaux;
      } else {
        currentchatdivid="home_container";
      }
      if (last_rend == "settings") {
        lastchatdivid="settings";
      } else if (last_rend != "")  {
        lastchatdivid="chat" + last_rend;
      } else {
        lastchatdivid="home_container";
      }
      let currentchatdiv = document.getElementById(currentchatdivid);
      if (currentchatdiv) {
        currentchatdiv.style.display = "block";
      }
      let lastchatdiv = document.getElementById(lastchatdivid);
      if (lastchatdiv) {
        lastchatdiv.style.display = "none";
      }
      SetUsers(Users, current_redaux, directmessages);
      let nowchatoptionsbutton = document.getElementById(
        "chatoptions" + current_redaux
      );
      if (nowchatoptionsbutton) {
        nowchatoptionsbutton.style.border = "1px solid #fff";
      }
      let lastchatoptionsbutton = document.getElementById(
        "chatoptions" + last_rend
      );
      if (lastchatoptionsbutton) {
        lastchatoptionsbutton.style.border = "none";
      }
     
      return;
    };

    export   async function showsendBtn(Files,current_red) {
      var sendBtn = document.getElementById("sendBtn" + current_red);
      var textarea = document.getElementById("inputtextarea" + current_red);
      
      if (textarea.value.trim() !== "" || Files[current_red].length > 0) {
        sendBtn.style.opacity = "100%";
        sendBtn.style.pointerEvents = "auto";
      } else {
        sendBtn.style.opacity = "0%";
        sendBtn.style.pointerEvents = "none";
      }
    }