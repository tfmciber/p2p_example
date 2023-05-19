<script>
  import { getColorForUserId } from "./utils.js";
  import { checkPasswordMatch } from "./utils.js";
  import { createCommand } from "./utils.js";
  import { updateProgress } from "./utils.js";
  import { SetUsers } from "./utils.js";
  import { createMessage } from "./utils.js";
  import {auxchangechat} from "./utils.js";
  import {showsendBtn} from "./utils.js";
  import { addRend } from "./rend.js";
  import { reload } from "./rend.js";
  import { cancelRendezvousstr } from "./rend.js";
  
  import { NewHost } from "../wailsjs/go/main/P2Papp.js";
  import { RestartApplication } from "../wailsjs/go/main/P2Papp.js";
  import { ReadKeys } from "../wailsjs/go/main/P2Papp.js";
  import { NewID } from "../wailsjs/go/main/P2Papp.js";
  import { OpenID } from "../wailsjs/go/main/P2Papp.js";
  import { Clear } from "../wailsjs/go/main/P2Papp.js";
  import {HostStats} from "../wailsjs/go/main/P2Papp.js";
  import {ChangePassword} from "../wailsjs/go/main/P2Papp.js";
  import {DeleteAccount} from "../wailsjs/go/main/P2Papp.js";
  import {LoadData} from "../wailsjs/go/main/P2Papp.js";
  import {SendDM} from "../wailsjs/go/main/P2Papp.js";


  import { SendTextHandler } from "../wailsjs/go/main/P2Papp.js";
  import { SelectFiles } from "../wailsjs/go/main/P2Papp.js";
  import { QueueFile } from "../wailsjs/go/main/P2Papp.js";

  import { LeaveChat } from "../wailsjs/go/main/P2Papp.js";
  import { DeleteChat } from "../wailsjs/go/main/P2Papp.js";
  import uploadBtn from "./assets/images/uploadBtn.png";
  import fileIcon from "./assets/images/folder.png";
  import micBtn from "./assets/images/microphone.png";
  import pauseBtn from "./assets/images/pause.png";

  import Chart from 'chart.js/auto'

  let current_red = "";
  let ciphered = [];
  let id = "";
  
  
  let Users = {};
  let Files = {};
  var chats = {};
 
  var directmessages = {};
  let filename = "key.key";
  let password = "";
  let login_register = true;
  let loggedin = false;
  
  let sysMemorychart;
  let sysNumFDchart;
  let sysNumConnschart;
  let sysNumStreamschart;

  let transMemorychart;
  let transNumFDchart;
  let transNumConnschart;
  let transNumStreamschart;


  async function startup() {
    
    await ReadKeys(filename).then((result) => (ciphered = result));
    
    loggedin = false;
  
    await Clear().then();
    if (ciphered == null) {
      login_register = false;
    } else {
      login_register = true;
    }
 
    current_red = "";
    
    chats= {};
    Users={};
    Files = {};
    directmessages = {};
  }
  startup();


function startGraphs(){


  let sysMemorychartcanvas = document.getElementById("sysMemorychart")
  let sysNumFDchartcanvas = document.getElementById("sysNumFDchart")
  let sysNumConnschartcanvas = document.getElementById("sysNumConnschart")
  let sysNumStreamschartcanvas = document.getElementById("sysNumStreamschart")

 
  let transMemorychartcanvas = document.getElementById("transMemorychart")
  let transNumFDchartcanvas = document.getElementById("transNumFDchart")
  let transNumConnschartcanvas = document.getElementById("transNumConnschart")
  let transNumStreamschartcanvas = document.getElementById("transNumStreamschart")



    
     sysMemorychart = createGraph1(sysMemorychartcanvas,"System Memory Usage");
      sysNumFDchart = createGraph1(sysNumFDchartcanvas,"System File Descriptors");
      sysNumConnschart = createGraph2(sysNumConnschartcanvas,"System Connections","In","Out");
      sysNumStreamschart = createGraph2(sysNumStreamschartcanvas,"System Streams","In","Out");

   transMemorychart = createGraph1(transMemorychartcanvas,"Transient Memory Usage");
   transNumFDchart = createGraph1(transNumFDchartcanvas,"Transient File Descriptors");
   transNumConnschart = createGraph2(transNumConnschartcanvas,"Transient Connections","In","Out");
      transNumStreamschart = createGraph2(transNumStreamschartcanvas,"Transient Streams","In","Out");


  }
 
function createGraph1(canvas,Name){
  if (canvas == null) {
    return;
  }
  let ctx = canvas.getContext('2d');
  let chart = new Chart(ctx, {
    type: 'line',
    
    data: {
      datasets: [{
        label: Name,
        data: [],
        backgroundColor: 'rgb(173, 216, 230)',
        borderColor: 'rgb(173, 216, 230)',
        borderWidth: 1,
        pointRadius: 0,
        fill: false,
        tension: 0.1
      }]
    },
    options:  {
      
      maintainAspectRatio: false,
      plugins: {
            title: {
                display: true,
                text: Name
            },
            legend: {
    display: false
  },
        },
			scales: {
				xAxes: [{
					type: 'time',
					time: {
						unit: 'hour',
						displayFormats: {
							hour: 'HH:mm:ss'
						}
					}
				}],
				yAxes: [{
					scaleLabel: {
						display: true,
						labelString: 'My Y-Axis Label'
					}
				}]
			}
}
		
  });
  return chart;
}

function createGraph2(canvas,Title,Name1,Name2){
  if (canvas == null) {
    return;
  }
  let ctx = canvas.getContext('2d');
  let chart = new Chart(ctx, {
    type: 'line',
    
    data: {
      datasets: [{
        label: Name1,
        data: [],
        backgroundColor: 'rgb(0, 128, 0)',
        borderColor: 'rgb(0, 128, 0)',
        borderWidth: 1,
        pointRadius: 0,
        fill: false,
        tension: 0.1
      },
      {
        label: Name2,
        data: [],
        backgroundColor: 'rgb(173, 216, 230)',
        borderColor: 'rgb(173, 216, 230)',
        borderWidth: 1,
        pointRadius: 0,
        fill: false,
        tension: 0.1
      }
    ]
    },
    options:  {
      
      maintainAspectRatio: false,
      plugins: {
            title: {
                display: true,
                text: Title
            },
            
        },
			scales: {
				xAxes: [{
					type: 'time',
					time: {
						unit: 'hour',
						displayFormats: {
							hour: 'HH:mm:ss'
						}
					}
				}],
				yAxes: [{
					scaleLabel: {
						display: true,
						labelString: 'My Y-Axis Label'
					}
				}]
			}
}
		
  });
  return chart;
}
function addData1(chart,x, y) {
  chart.data.labels.push(x);
  chart.data.datasets[0].data.push(y);
  chart.update();
  }
  function addData2(chart,x, y1,y2) {
  chart.data.labels.push(x);
  chart.data.datasets[0].data.push(y1);
  chart.data.datasets[1].data.push(y2);
  chart.update();
  }


  function login() {
    ciphered = document.getElementById("ciphered").value;
    password = document.getElementById("password").value;
    OpenID(ciphered, password).then((result) => test1(result));
  }
  function test1(result) {
    {
      if (result == "") {
        //show error message in html
        startHost().then((result) => (loggedin = result));
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

  function register() {
   
    password = document.getElementById("password").value;
    NewID(password, filename).then();
    startHost().then((result) => (loggedin = result));
    


  }
  function loadData(){
    LoadData().then();
  }
  async function startHost() {
    await NewHost().then((result) => (id = result));
      
 
    HostStats().then();

    setTimeout(loadData,0);
    
    setTimeout(startGraphs,0);
    

    return true;
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
      

      await SetUsers(Users, current_red, directmessages);
    });
  }




  function DMleft() {
    window.runtime.EventsOn("dmLeft", async function () {
      await ChangeChat("");
    });
  }
function Statistics(){
    window.runtime.EventsOn("Statistics", async function (stats) {
      
      let date = new Date().toLocaleTimeString();
      addData1(sysMemorychart,date, stats.sysMemory);
      addData1(sysNumFDchart,date, stats.sysNumFD);
      addData2(sysNumConnschart,date, stats.sysNumConnsInbound, stats.sysNumConnsOutbound);
      addData2(sysNumStreamschart,date, stats.sysNumStreamsInbound, stats.sysNumStreamsOutbound);

      addData1(transMemorychart,date, stats.transMemory);
      addData1(transNumFDchart,date, stats.transNumFD);
      addData2(transNumConnschart,date, stats.transNumConnsInbound, stats.transNumConnsOutbound);
      addData2(transNumStreamschart,date, stats.sysNumStreamsInbound, stats.sysNumStreamsOutbound);
      

      
    });
  }

  function receiveMessage() {
    window.runtime.EventsOn(
      "receiveMessage",
      function (arg1, arg2, arg3, arg4) {

        createMessage(arg1, arg2, arg3, arg4, [],false);
      }
    );
  }

  function loadMessages() {
    window.runtime.EventsOn(
      "loadMessages",
      function (arg1, arg2, arg3, arg4,arg5,arg6) {
        createMessage(arg1, arg2, arg3, arg4, arg5, arg6);
        
      }
    );
  }

  function seachRend() {
    window.runtime.EventsOn("searchRend", function (arg) {

      let chatbuttondiv = document.getElementById("chatoptions"+arg);
      
      if ( chatbuttondiv!=null){  
        
        //get button in div with class chatoption
        let but = chatbuttondiv.getElementsByClassName("chatoptions")[0];

        let reloadbtn = chatbuttondiv.getElementsByClassName("reloadchatbtn")[0];
        let cancelbtn = chatbuttondiv.getElementsByClassName("removechatbtn")[0];
        cancelbtn.style.display = "inline";
        reloadbtn.style.display = "none";
            
        but.className = "chatoptions-loading"
  }
   
    });
  }


  function endRend() {
    window.runtime.EventsOn("endRend", function (arg) {
     
      let chatbuttondiv = document.getElementById("chatoptions"+arg);
      
      if ( chatbuttondiv!=null){  
        
        //get button in div with class chatoption
        let but = chatbuttondiv.getElementsByClassName("chatoptions-loading")[0];

        let reloadbtn = chatbuttondiv.getElementsByClassName("reloadchatbtn")[0];
        let cancelbtn = chatbuttondiv.getElementsByClassName("removechatbtn")[0];
        cancelbtn.style.display = "none";
        reloadbtn.style.display = "inline";
            
        but.className = "chatoptions"
  }
    });
  }
  seachRend();
  endRend();

  
  function direcMessage() {
    window.runtime.EventsOn("directMessage", function (arg) {
      directmessages = arg;
    });
  }
  function receiveFile() {
    window.runtime.EventsOn("receiveFile", async function (...arg) {
      await createMessage(arg[0], "", arg[1], "", [arg[2]],false);
    });
  }
  function terminal() {
    window.runtime.EventsOn("receiveCommands", function (...arg) {
      createCommand(arg);
    });
  }
  async function progressFile() {
    window.runtime.EventsOn("progressFile", async function (...arg) {
      await updateProgress(arg[0], arg[1], arg[2], arg[3],arg[4]);
    });
  }


  loadMessages();


  terminal();
  receiveMessage();
  receiveFile();
  updateUsers();
  updateChats();
  direcMessage();
  progressFile();
  DMleft();
  Statistics();




async function deleteAccount() {


  await  DeleteAccount(filename).then((result)=>
  {
    if (result == false){
      
      let spandi = document.getElementById("account-delete-error");
      spandi.style.display = "block";
    }

  }
  
  );
  startup();


  }

   async function changePassword(){

    let currentpassword=document.getElementById("currentpassword").value;
    let newpassword=document.getElementById("newpassword").value;

    
    
    document.forms["changepassword"].reset();

    if (ciphered == null){
      await ReadKeys(filename).then((result) => (ciphered = result));
    }

    ChangePassword(currentpassword,newpassword,ciphered,filename).then((result)=> async function()
    {
      if (result == true){

        let spandi = document.getElementById("password-change-success");
        spandi.style.display = "block";
        await ReadKeys(filename).then((result) => (ciphered = result));
        //set style to display none after 5 seconds
        setTimeout(function(){ spandi.style.display = "none"; }, 5000);
      }
      else{
        
        let spandi = document.getElementById("password-change-error");
        spandi.style.display = "block";
      }
    
    }
    );
    
  }

  async function sendmessage(message, setmsg,dest) {

    if (setmsg != true) {
      message = document.getElementById("inputtextarea" + dest).value;
      let input = document.getElementById("inputtextarea" + dest);
      var sendBtn = document.getElementById("sendBtn" + dest);

      input.value = "";
      sendBtn.style.opacity = "0%";
      sendBtn.style.pointerEvents = "none";
    }

    if (message != "") {
      await SendTextHandler(message, dest).then((result) => {
        createMessage(dest, message, "me", new Date().toLocaleString(), "", result);
      });
    }

    if (Files[dest] != null && Files[dest].length > 0) {
      createMessage(dest, "", "me", new Date().toLocaleString(), Files[dest], false);
      for (let i = 0; i < Files[dest].length; i++) {
        let file = Files[dest][i];
        let path = file.path;

        QueueFile(dest, path).then();
      }

      Files[dest] = [];

      //clear files in html
      let files = document.getElementById("filescontainer" + dest);
      files.innerHTML = "";

    files.style.display = "none";

  
    }
  }

  async function leaveChat(arg) {
    await LeaveChat(arg).then();
   
    setTimeout(ChangeChat,0,"");
 
  }
  async function deleteChat(arg) {
    
    await DeleteChat(arg).then();
    setTimeout(ChangeChat,0,"");
  }

  async function addfile() {
    //files is array of struct of path and size
    let newfiles = [];
    const container = document.getElementById("filescontainer" + current_red);
    container.style.display = "flex";

    await SelectFiles().then((result) => {
      result.forEach((pathfilename) => {
        let path = pathfilename.path;

        let filename = pathfilename.filename;
        let progress = pathfilename.progress;
        if (!Files[current_red]) {
          Files[current_red] = [];
        }

        if (!Files[current_red].find((file) => file.path === path)) {
          Files[current_red].push({ path, filename,progress });
          newfiles.push({ path, filename ,progress});
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
        if (Files[current_red].length == 0){
          container.style.display = "none";
        }
        showsendBtn(Files,current_red);
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
    showsendBtn(Files,current_red);
  }
  //Manually disable pinch zoom!
  document.addEventListener(
    "wheel",
    (event) => {
      const { ctrlKey } = event;
      if (ctrlKey) {
        event.preventDefault();
        return;
      }
    },
    { passive: false }
  );
  async function ChangeChat(chat) {
    
   
    if (chat == current_red) {
      return;
    }
    

    
    auxchangechat(current_red,chat,Users,directmessages);
   
    current_red = chat;

    if (chat != "settings" && chat != ""){
    if (Files[chat] == null) {
      Files[chat] = [];
    }
    //scroll chats-menu to the current chat button
    let chatsmenu = document.getElementById("chats-menu");
    let button = document.getElementById("chatoptions" + chat);
    let scrollHeight = button.offsetTop - 100;
    chatsmenu.scrollTo(0, scrollHeight);
    }
  }


  function deleteFile(element) {
    for (let i = 0; i < Files[current_red].length; i++) {
      if (Files[current_red][i].filename == element) {
        Files[current_red].splice(i, 1);
        break;
      }
    }
  }

  async function textareacheck() {
    var textarea = document.getElementById("inputtextarea" + current_red);

    textarea.style.height = "16px";
    textarea.style.height = `${Math.min(textarea.scrollHeight, 56)}px`;

    await showsendBtn(Files,current_red);
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

  function getContrastColor(color){
    return (parseInt(color.replace("#",""), 16) > 0xffffff/2) ? 'black':'white';
  }

  async function sendAndAddtoChat() {
    var modal = document.getElementById("popup");
    var poputname = modal.querySelector("#popupname");

    modal.style.display = "none";

    let auxdirectmessages = directmessages;
    if (!directmessages.find((ch) => ch === poputname.innerText)) {
      auxdirectmessages.push(poputname.innerText);
    }
    SendDM(poputname.innerText).then();
    directmessages = auxdirectmessages;
   
   var mess = document.getElementById("textinpopup");

  
    sendmessage(mess.value, true, poputname.innerText);
    document.forms["chatinpopup"].reset();
   setTimeout(ChangeChat,0,poputname.innerText);
    
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
        <div class="chats-menu" id ="chats-menu">
       
          {#each Object.entries({...chats, ...directmessages})  as [chat, item]}
          {#if chat != ""}
         
          <div  id="chatoptions{chat}">
            {#if item.Status == true}
            <button
              type="button"
              class="chatoptions"
              style="background-color: {getColorForUserId(chat)};color: {getContrastColor(getColorForUserId(chat))}"
              on:click={() => ChangeChat(chat)}
            >
              {chat}</button
            >
            {:else}
            <button
              type="button"
              class="chatoptions"
              on:click={() => ChangeChat(chat)}
            >
              {chat}</button
            >
            {/if}
            <button class="reloadchatbtn" on:click={() => reload(chat)}>&#x21bb;</button>
            <button  class ="removechatbtn" on:click={() => cancelRendezvousstr(chat)}> &#x2715 </button>
          </div>
          {/if}
          {/each}
          
        </div>
        <div class="option">
          <button on:click={() => RestartApplication().then()}>
            Log out
            <i class="fas fa-sign-out-alt" />
          </button>
          <button on:click={() => ChangeChat("settings")}> Settings </button>
        </div>
      </div>

      <div class="data-container">
        <div id="settings">
<h1> Settings</h1>
<table class="settingstable">
  <tr>
    <td>Host ID</td>
    <td>{id}</td>
  </tr> 4n


  <tr>
    <td>Change password</td>
    <td>
      <form
      class = "settingsform"
        autocomplete="off"
        id="changepassword"
        on:input={() => checkPasswordMatch("#newpassword", "#confirmnewpassword", "#password-match-error2","changepassword")}
        on:submit|preventDefault={changePassword}
        
      >
      <label for="currentpassword">Current Password</label>
        <input
          type="password"
          placeholder="Enter current password"
          id="currentpassword"
          class = "settingsinput"
          name="currentpassword"
          required
        />
        <label for="newpassword">New Password</label>
        <input
          type="password"
          placeholder="Enter new password"
          id="newpassword"
          
          name="newpassword"
          required
        />
        <label for="confirmnewpassword">Confirm New Password</label>
        <input
          type="password"
          placeholder="Confirm new password"
          id="confirmnewpassword"
          
          name="confirmnewpassword"
          required
        />
        
        <button id="submit-btn" class="submit-btn"  type="submit" > Change </button>
       
      <span id="password-change-error" style="color: red; display: none; top:57%"
        >Error Changing Password</span
      >

      <span id="password-change-success" style=" color: green; display: none; top:28%;left: 20%"
      >Password changed Succesfully</span
    >
      <span id="password-match-error2" style="color: red; display: none; top:57%"
      >Passwords do not match</span
    >
      </form>   
    </td>
  </tr>
  <tr>
    <td>Delete account</td>

    <td> 
      
      <form
      autocomplete="off"
      id="changepassword"
      
      on:submit|preventDefault={deleteAccount}
    >
      <button
        type="submit"
        disabled
        style="display: none"
        aria-hidden="true"
      />
      <input
        type="password"
        placeholder="Enter current password"
        id="currentpassword"
        
        name="currentpassword"
        required
      />

      <button id="submit-btn" class="submit-btn"> Delete Account </button>
    </form>
    <span id="account-delete-error" style="color: red; display: none; top:57%"
    >Error deleting account</span
  >
    </td>
  </tr>
</table>
<br>
<h1>Advanced</h1>
<div id="terminal">
  Terminal
  <div id="terminal-box" />
</div>
<br>
  <h1>Network Statistics</h1>

<br>

  <h2>System Statistics</h2>

<div class= "chartsboxes">
  

  <div class = "chart" >
  <canvas id="sysMemorychart"></canvas>
</div>
<div class = "chart" >
  <canvas id="sysNumFDchart"></canvas>
</div>
<div class = "chart" >
  <canvas id="sysNumConnschart"></canvas>
</div>
<div class = "chart" >
  <canvas id="sysNumStreamschart"></canvas>
</div>
  

</div>
<h2>Transient Statistics</h2>

<div class= "chartsboxes">
   
<div class = "chart" >


  <canvas id="transMemorychart"></canvas>
</div>
<div class = "chart" >
  <canvas  id="transNumFDchart"></canvas>
</div>
<div class = "chart" >

  <canvas  id="transNumConnschart"></canvas>
</div>
<div class = "chart" >
  <canvas  id="transNumStreamschart"></canvas>
</div>

</div>

</div>
        <div id="home_container" class="homecontainer">
          <h4>Host ID: {id}</h4>

          <div class="rend-container">
            <div class="rendform">
              <form
                autocomplete="off"
                id="rendform"
                on:submit|preventDefault={addRend}
              >
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

            </div>
           
          </div>

        </div>

        <div class="chatdiv">
         
          {#each Object.entries({...chats, ...directmessages})  as [chat, item]}
          {#if chat != ""}
          <div class="chatdiveach" id="chat{chat}">
            {#if item.Status == true}
              <button
                class="leave-chat"
                id="buttonleave{chat}"
                title="Leave chat"
                on:click={() => leaveChat(chat)}
              >
              &#x2715
              </button>
            {:else}
              <button
                class="leave-chat"
                id="buttonleave{chat}"
                title="Delete chat"
                on:click={() => deleteChat(chat)}
              >
               &#xF5DE
              </button>
            {/if}

            <h1 class="chatname">{chat}</h1>
            
              <div class="chat-box" id="chat-box{chat}">
                <div class="filecontainers" id="filescontainer{chat}" />
              </div>

              {#if item.Status == true}
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
                    on:click={() => sendmessage(null,null,chat)}
                  
                    />

                </div>
                {/if}
            </div>
            {/if}
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
              required
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
          <form class="login-form"  id="login-form" on:submit|preventDefault={register}
          
          on:input={() => checkPasswordMatch("#password", "#confirm-password", "#password-match-error","login-form")}
          >
            <label for="password">Password</label>
            <input type="password" id="password" name="password" />

            <label for="confirm-password">Confirm Password</label>
            <input
              type="password"
              id="confirm-password"
              name="confirm-password"
              
              required
            />
            <button type="submit">Register</button>
            <span id="password-match-error" style="position:absolute; color: red; display: none; top:57%"
              >Passwords do not match</span
            >
           
          </form>
          <p>Already have an account?</p>
          <button on:click={() => (login_register = true)}>Login</button>
        {/if}
      </div>
    {/if}
  </div>
</body>
