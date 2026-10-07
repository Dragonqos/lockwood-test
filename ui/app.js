(() => {
  const state = {
    socket: null,
    requestID: 0,
    user: "",
  };

  const $ = (id) => document.getElementById(id);
  const connectionState = $("connectionState");
  const connectButton = $("connectButton");
  const messages = $("messages");
  const log = $("log");

  function writeLog(message, data) {
    const suffix = data === undefined ? "" : ` ${JSON.stringify(data)}`;
    log.textContent += `${new Date().toLocaleTimeString()} ${message}${suffix}\n`;
    log.scrollTop = log.scrollHeight;
  }

  function setConnectionState(label, connected) {
    connectionState.textContent = label;
    connectionState.className = `state ${connected ? "online" : "offline"}`;
    connectButton.textContent = connected ? "Disconnect" : "Connect";
  }

  function nextRequestID() {
    state.requestID += 1;
    return state.requestID;
  }

  function send(command) {
    if (!state.socket || state.socket.readyState !== WebSocket.OPEN) {
      writeLog("Not connected");
      return false;
    }
    state.socket.send(JSON.stringify(command));
    writeLog("→", command);
    return true;
  }

  function connect() {
    const user = $("user").value.trim();
    if (!user) return;

    const protocol = location.protocol === "https:" ? "wss:" : "ws:";
    state.user = user;
    state.socket = new WebSocket(`${protocol}//${location.host}/ws`);
    setConnectionState("Connecting...", false);

    state.socket.addEventListener("open", () => {
      setConnectionState("Connected", true);
      writeLog("Connected");
      send({ rID: nextRequestID(), cmd: "auth", user });
    });

    state.socket.addEventListener("message", (event) => {
      let payload;
      try {
        payload = JSON.parse(event.data);
      } catch {
        writeLog("← invalid JSON", event.data);
        return;
      }

      writeLog("←", payload);
      const isRoomEvent = !Object.prototype.hasOwnProperty.call(payload, "rID");
      if (isRoomEvent && (payload.cmd === "chat" || payload.cmd === "user_joined" || payload.cmd === "user_left")) {
        addMessage(payload);
      }
    });

    state.socket.addEventListener("close", () => {
      setConnectionState("Offline", false);
      writeLog("Disconnected");
      state.socket = null;
    });

    state.socket.addEventListener("error", () => writeLog("WebSocket error"));
  }

  function disconnect() {
    if (state.socket) state.socket.close();
  }

  function addMessage(event) {
    const item = document.createElement("div");
    item.className = `message ${event.cmd === "chat" ? "chat-message" : "system-message"}`;
    if (event.cmd === "chat") {
      item.innerHTML = `<strong></strong><span></span>`;
      item.querySelector("strong").textContent = event.user || "unknown";
      item.querySelector("span").textContent = event.message || "";
    } else {
      item.textContent = `${event.user || "Someone"} ${event.cmd === "user_joined" ? "joined" : "left"}`;
    }
    messages.appendChild(item);
    messages.scrollTop = messages.scrollHeight;
  }

  $("authForm").addEventListener("submit", (event) => {
    event.preventDefault();
    if (state.socket) disconnect();
    else connect();
  });

  $("roomForm").addEventListener("click", (event) => {
    const command = event.target.dataset.command;
    if (!command) return;
    const name = $("room").value.trim();
    if (name) send({ rID: nextRequestID(), cmd: command, name });
  });

  $("leaveButton").addEventListener("click", () => {
    send({ rID: nextRequestID(), cmd: "leave_room" });
  });

  $("chatForm").addEventListener("submit", (event) => {
    event.preventDefault();
    const input = $("message");
    const message = input.value.trim();
    if (message && send({ rID: nextRequestID(), cmd: "chat", message })) {
      input.value = "";
    }
  });

  $("clearLogButton").addEventListener("click", () => { log.textContent = ""; });
})();
