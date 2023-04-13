<script>
  import { NewHost } from "../wailsjs/go/main/P2Papp.js";
  import { ReadKeys } from "../wailsjs/go/main/P2Papp.js";
  import { NewID } from "../wailsjs/go/main/P2Papp.js";
  import { OpenID } from "../wailsjs/go/main/P2Papp.js";
  import { Clear } from "../wailsjs/go/main/P2Papp.js";
  import { InitDHT } from "../wailsjs/go/main/P2Papp.js";
  import { DhtRoutine } from "../wailsjs/go/main/P2Papp.js";
  import { AddRendezvous } from "../wailsjs/go/main/P2Papp.js";
  import { CancelRendezvous } from "../wailsjs/go/main/P2Papp.js";
  import { SendTextHandler } from "../wailsjs/go/main/P2Papp.js";
  import { SelectFiles } from "../wailsjs/go/main/P2Papp.js";
  import { SendFile } from "../wailsjs/go/main/P2Papp.js";
  import { OpenFileExplorer } from "../wailsjs/go/main/P2Papp.js";


  import uploadBtn from "./assets/images/uploadBtn.png";

  import fileIcon from "./assets/images/folder.png";
  import check from "./assets/images/check.png";
  import wrong from "./assets/images/wrong.png";
  /*
TODO: BUTTON TO DELETE A chat
todo show progress when uploading and downloading a file

*/
  let current_red = "";
  let ciphered = [];
  //var of type number[]
  let id = "";
  let Users = {}; // Users[""] will hold all users

  let Files = {};
  var chats = [];
  var directmessages = [];
  let filename = "key.key";
  let password = "";
  let login_register = true; // true for login, false for register
  let loggedin = false;

  async function startup() {
    //check if file key.key exists, if exists login_register = true, else login_register = false

    await ReadKeys(filename).then((result) => (ciphered = result));
    loggedin = false;
    await Clear().then();
    if (ciphered == null) {
      login_register = false;
    } else {
      login_register = true;
    }
    current_red = "";
    chats = [];
    Users = {};
    Files = {};
    directmessages = [];
  }
  startup();

  function login() {
    ciphered = document.getElementById("ciphered").value;
    password = document.getElementById("password").value;
    OpenID(ciphered, password).then((result) => test1(result));
  }
  function test1(result) {
    {
      if (result == "") {
        //show error message in html
        startHost();
      } else {
        //change ciphered item in html to text in red
        let text = "~ Id or password are not correct";
        document.getElementById("idlabel").innerHTML = "Id " + text;
        document.getElementById("idlabel").style.color = "red";
        document.getElementById("plabel").innerHTML = "Password " + text;
        document.getElementById("plabel").style.color = "red";
      }
    }
  }
  async function SetUsers() {
    let usersAux = Users[current_red];
    if (directmessages.includes(current_red)) {
      usersAux = Users[""].filter((user) => user.user == current_red);
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
  }

  function register() {
    password = document.getElementById("password").value;
    NewID(password, filename).then();
    startHost();
  }
  async function startHost() {
    loggedin = true;
    await NewHost().then((result) => (id = result));
    //wait for host to be created

    await InitDHT();
    DhtRoutine(true).then();
    
  }

  function updateChats() {
    window.runtime.EventsOn("updateChats", function (arg) {
      chats = arg;
    });
  }
  function updateUsers() {
    window.runtime.EventsOn("updateUsers", async function (arg) {
      
        
      Users = {};
      for (let i = 0; i < arg.length; i++) {
        Users[arg[i].chat] = [];
        for (let j = 0; j < arg[i].user.length; j++) {
        
          let user = {
            user: arg[i].user[j].ip,
            online: arg[i].user[j].status,
          };
          Users[arg[i].chat].push(user);
        }
      }
     
      await SetUsers();
    });
  }
  async function receiveMessage() {
    window.runtime.EventsOn("receiveMessage", async function (...arg) {
      await createMessage(arg[0], arg[1], arg[2], arg[3]);
    });
  }
  function direcMessage() {
    window.runtime.EventsOn("directMessage", function (arg) {
      directmessages = arg;
    });
  }
  function receiveFile() {
    window.runtime.EventsOn("receiveFile", function (...arg) {
      createMessage(arg[0], arg[1], arg[2]);
    });
  }
  function terminal() {
    window.runtime.EventsOn("receiveCommands", function (...arg) {
      createCommand(arg);
    });
  }
  terminal();
  receiveMessage();
  receiveFile();
  updateUsers();
  updateChats();
  direcMessage();

  function cancelRendezvous() {
    CancelRendezvous().then();
  }

  async function addRend() {
    const loader = document.querySelector(".loader");
    const submitBtn = document.getElementById("submit-btn");
    const cancelBtn = document.getElementById("cancel-btn");
    let rend = document.getElementById("rend").value;

    //show loader
    loader.style.display = "block";
    cancelBtn.style.display = "block";
    submitBtn.style.display = "none";
    await AddRendezvous(rend).then();
    loader.style.display = "none";
    cancelBtn.style.display = "none";
    submitBtn.style.display = "";
  }
  function createCommand(cmd) {
    let terminal = document.getElementById("terminal-box");
    let command = document.createElement("div");
    command.className = "command";
    command.innerHTML = cmd;
    terminal.scrollTop = terminal.scrollHeight + 20;
    terminal.appendChild(command);
  }
  async function sendmessage(message, setmsg) {
    if (setmsg != true) {
      message = document.getElementById("inputtextarea" + current_red).value;
      let input = document.getElementById("inputtextarea" + current_red);
      var sendBtn = document.getElementById("sendBtn" + current_red);

      input.value = "";
      sendBtn.style.opacity = "0%";
      sendBtn.style.pointerEvents = "none";
    }

    await SendTextHandler(message, current_red).then((result) => {
      createMessage(current_red, message, "me", "", Files[current_red], result);
    });

    if (Files[current_red] != null) {
      for (let i = 0; i < Files[current_red].length; i++) {
        let file = Files[current_red][i];
        let path = file.path;

        SendFile(current_red, path).then();
      }

      Files[current_red] = [];

      //clear files in html
      let files = document.getElementById("filescontainer" + current_red);
      files.innerHTML = "";
    }
  }

  //creates a new message div and returns it
  function createMessage(chat, message, sender, time, files, ok) {
    let chatbox = document.getElementById("chat-box" + chat);
    let newmessage = document.createElement("div");
    if (sender == "me") {
      newmessage.className = "messagesent";
      //get time and date dd/mm/yyyy hh:mm
      time = new Date().toLocaleString();
    } else {
      newmessage.className = "messagereceived";
    }
    //if files are not null, add them to the message

    let container = document.createElement("div");
    container.className = "fileIconmessageSCONTAINER";
    if (files != null) {
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
        container.appendChild(button);
        container.appendChild(text);
      }
    }

    newmessage.innerHTML = `<div class="message-header">
    <div class="message-sender">${sender}</div>
    <div class="message-time">${time}</div>
  </div>
    <div class="message-text">${message}</div>
    `;
    if (sender == "me" && message != "") {
      if (ok == true) {
        newmessage.innerHTML += `<img class="message-ok" src=${check} alt="ok" />`;
      } else {
        newmessage.innerHTML += `<img class="message-ok" src=${wrong} alt="ok" />`;
      }
    }
    if (files != null) {
      newmessage.appendChild(container);
    }

    chatbox.appendChild(newmessage);
    //scroll to bottom
    chatbox.scrollTop = chatbox.scrollHeight;
  }

  function checkPasswordMatch() {
    const passwordInput = document.querySelector("#password");
    const confirmPasswordInput = document.querySelector("#confirm-password");
    const passwordMatchError = document.querySelector("#password-match-error");

    if (passwordInput.value !== confirmPasswordInput.value) {
      passwordMatchError.style.display = "block";
    } else {
      passwordMatchError.style.display = "none";
    }
  }

  async function addfile() {
    //files is array of struct of path and size
    let newfiles = [];
    const container = document.getElementById("filescontainer" + current_red);

    await SelectFiles().then((result) => {
      result.forEach((pathfilename) => {
        let path = pathfilename.path;

        let filename = pathfilename.filename;
        if (!Files[current_red]) {
          Files[current_red] = [];
        }

        if (!Files[current_red].find((file) => file.path === path)) {
          Files[current_red].push({ path, filename });
          newfiles.push({ path, filename });
        }
      });
    });

    for (let i = 0; i < newfiles.length; i++) {
      let icon = document.createElement("img");
      icon.className = "fileicon";
      //icon image depends on file type, if image then show image, else show a default icon
      let filename = newfiles[i].filename;

      icon.src = fileIcon;
      let filediv = document.createElement("div");
      //on mouse over show button to remove file, on mouse out hide button

      const button = document.createElement("button");
      // Add a click event listener to the button
      button.addEventListener("click", () => {
        const name =
          button.parentElement.querySelector(".filedivname").innerText;
        button.parentElement.querySelector(".filedivname").parentNode.remove();
        deleteFile(name);
        showsendBtn();
      });
      button.innerHTML = '<i class="fas fa-trash"></i>';
      button.className = "removefilebtn";

      filediv.className = "filediv";
      filediv.innerHTML = `
          <div class="filedivname">${filename}</div>
          ${icon.outerHTML}`;

      filediv.appendChild(button);
      container.appendChild(filediv);
    }
    showsendBtn();
  }

  async function ChangeChat(chat) {
    if (chat == current_red) {
      return;
    }

    if (Files[chat] == null) {
      Files[chat] = [];
    }
    setTimeout(auxchangechat, 100, current_red);

    current_red = chat;
    setTimeout(SetUsers, 100);
  }

  async function auxchangechat(last_rend) {
    if (current_red != "") {
      let currentchatdiv = document.getElementById("chat" + current_red);
      currentchatdiv.style.display = "block";
    } else {
      let home_container = document.getElementById("home_container");
      home_container.style.display = "block";
    }
    if (last_rend != "") {
      let lastchatdiv = document.getElementById("chat" + last_rend);
      lastchatdiv.style.display = "none";
    } else {
      let home_container = document.getElementById("home_container");
      home_container.style.display = "none";
    }
    let nowchatoptionsbutton = document.getElementById(
      "chatoptions" + current_red
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
  }
  function deleteFile(element) {
    for (let i = 0; i < Files[current_red].length; i++) {
      if (Files[current_red][i].filename == element) {
        Files[current_red].splice(i, 1);
        break;
      }
    }
  }
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
  async function textareacheck() {
    var textarea = document.getElementById("inputtextarea" + current_red);

    textarea.style.height = "16px";
    textarea.style.height = `${Math.min(textarea.scrollHeight, 80)}px`;

    await showsendBtn();
  }
  async function showsendBtn() {
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

  function hidepopup() {
    //detect click outside of popup, if so hide popup

    window.onclick = function (event) {
      let modal = document.getElementById("popup");
      //check if click is inside popup

      if (
        event.target.className != "useronline" &&
        event.target.className != "useroffline" &&
        !modal.contains(event.target)
      ) {
        modal.style.display = "none";
        return;
      }
      return;
    };
  }
  hidepopup();

  function sendAndAddtoChat() {
    var modal = document.getElementById("popup");
    var poputname = modal.querySelector("#popupname");

    modal.style.display = "none";

    let auxdirectmessages = directmessages;
    if (!directmessages.find((ch) => ch === poputname.innerText)) {
      auxdirectmessages.push(poputname.innerText);
    }
    directmessages = auxdirectmessages;
    ChangeChat(poputname.innerText);

    //change chat, update chatbox, update files

    setTimeout(SendMessageFrompopup, 100);
  }

  function SendMessageFrompopup() {
    var mess = document.getElementById("textinpopup");
    sendmessage(mess.value, true);
    document.forms["chatinpopup"].reset();
  }

  function getColorForUserId(userId) {
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
  }
</script>

<link
  rel="stylesheet"
  href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.15.3/css/all.min.css"
/>

<body>
  <div class="app-container">
    {#if loggedin}
      <div class="left-menu">
        <button class="Home" on:click={() => ChangeChat("")}> Home </button>

        {#each [...chats, ...directmessages] as chat}
          <button
            type="button"
            class="chatoptions"
            id="chatoptions{chat}"
            style="background-color: {getColorForUserId(chat)}"
            on:click={() => ChangeChat(chat)}
          >
            {chat}</button
          >
        {/each}
      </div>

      <div class="data-container">
        <div id="home_container">
          <h4>Host ID: {id}</h4>
          <h1>Join/Create channel</h1>

          <div class="rend-container">
            <div class="rendform">
              <form autocomplete="off" on:submit|preventDefault={addRend}>
                <button
                  type="submit"
                  disabled
                  style="display: none"
                  aria-hidden="true"
                />

                <input
                  type="text"
                  placeholder="Enter code"
                  id="rend"
                  class="rend-input"
                  name="rend"
                  required
                />
                <button id="submit-btn" class="submit-btn"> Join </button>
              </form>
              <button id="cancel-btn" on:click={() => cancelRendezvous()}>
                Cancel
              </button>
            </div>
            <div class="loader">
              <div class="dot-flashing" />
            </div>
            <button class="logout" on:click={() => startup()}>
              Log out
              <i class="fas fa-sign-out-alt" />
            </button>
          </div>
          <div id="terminal">
            Terminal
            <div id="terminal-box" />
          </div>
        </div>

        <div class="chatdiv" id="chatdiv">
          <h1 class="chatname">{current_red}</h1>
          {#each [...chats, ...directmessages] as chat}
            <div class="chatdiveach" id="chat{chat}">
              <div class="chat-box" id="chat-box{chat}">
                <div class="filecontainers" id="filescontainer{chat}" />
              </div>

              <div class="inputcontainer">
                <textarea
                  on:keyup={() => textareacheck()}
                  class="input-textarea"
                  id="inputtextarea{chat}"
                  placeholder="Send message ..."
                />
                <img
                  class="uploadlabed"
                  src={uploadBtn}
                  alt="img"
                  on:click={() => addfile()}
                />
                <input
                  type="file"
                  name="myfile"
                  id="file{chat}"
                  style="display:none"
                />

                <button
                  class="sendBtn"
                  id="sendBtn{chat}"
                  on:click={() => sendmessage()}
                />
              </div>
            </div>
          {/each}
        </div>
      </div>

      <div class="right-menu">
        <div id="popup">
          <div id="popupname" />

          <form
            class="chatinpopup"
            id="chatinpopup"
            autocomplete="off"
            on:submit|preventDefault={sendAndAddtoChat}
          >
            <input
              class="textinpopup"
              id="textinpopup"
              type="text"
              placeholder="Send Direct Message"
            />
            <button type="submit">Send</button>
          </form>
        </div>
        <h1>Users</h1>
        <div id="users">
          <div class="onlineusers" />
          <div class="offlineusers" />
        </div>
      </div>
    {:else}
      <div class="login-container">
        {#if login_register}
          <form
            class="login-form"
            autocomplete="off"
            on:submit|preventDefault={login}
          >
            <label for="ciphered" id="idlabel">Id</label>
            <input type="text" id="ciphered" value={ciphered} required />
            <label for="password" id="plabel">Password</label>
            <input type="password" id="password" required />
            <button class="btn">Login</button>
          </form>
          <p>
            Dont have an account? <button
              on:click={() => (login_register = false)}>Sign In</button
            >
          </p>
        {:else}
          <!-- Create a register form checking password confirmation matches  -->
          <form class="login-form" on:submit|preventDefault={register}>
            <label for="password">Password</label>
            <input type="password" id="password" name="password" />

            <label for="confirm-password">Confirm Password</label>
            <input
              type="password"
              id="confirm-password"
              name="confirm-password"
              on:input={checkPasswordMatch}
              required
            />
            <span id="password-match-error" style="color: red; display: none"
              >Passwords do not match</span
            >
            <button type="submit">Register</button>
          </form>
          <p>Already have an account?</p>
          <button on:click={() => (login_register = true)}>Login</button>
        {/if}
      </div>
    {/if}
  </div>
</body>
