const issuer = "https://{OKTA_ORG}.okta.com/oauth2/default";
const authClient = new OktaAuth({
    issuer: issuer
});

const windowFeatures = "toolbar=no,status=no,menubar=no,scrollbars=no,resizable=no, width=10, height=10, visible=none"
var windowObjectReference;

authClient.session.exists()
.then(function(exists) {
  if (exists) {
    // logged in
    authorize("");
  } else {
    // not logged in
    document.getElementById("loading").style.display = 'none';
    document.getElementById("okta-login-container").style.display = 'block';
  }
});

   

window.addEventListener("message", (event) => {
    // Do we trust the sender of this message?
    // if (event.origin !== "http://example.com:8080")
    //   return;
    document.getElementById("loading").style.display = 'none';
    document.getElementById("okta-login-container").style.display = 'block';
    console.log(event.origin);
    console.log(event.data);
    
    if (event.data !== undefined && !event.data.includes("auth_time")) {
        document.getElementById("loginError").innerHTML = `Authorization Error: ${event.data}`
        document.getElementById("loginError").style.display = 'block';
    } else {
        document.getElementById("loginSuccess").innerHTML = `Authorization Success: ${event.data}`
        document.getElementById("loginSuccess").style.display = 'block';
    }
   
    windowObjectReference.close();
}, false);


function login() {
    clearMessages();
    
    authClient.signInWithCredentials({
    username: document.getElementById("userName").value,
    password: document.getElementById("password").value
    })
    .then(function(transaction) {
    if (transaction.status === 'SUCCESS') {
        document.getElementById("userName").value = '';
        document.getElementById("password").value = '';
        authorize(transaction.sessionToken);
    } else {
        throw 'We cannot handle the ' + transaction.status + ' status';
    }
    })
    .catch(function(err) {
        document.getElementById("loginError").innerHTML = err;
        document.getElementById("loginError").style.display = 'block';
    });
}


function authorize(token) {
    const directAuthorizeUrl = `http://localhost:8082/authorize?session_token=${token}`
    windowObjectReference = window.open(directAuthorizeUrl, "loginI", windowFeatures);
}


function clearMessages() {
    document.getElementById("loginError").innerHTML = ''
    document.getElementById("loginError").style.display = 'none';
    document.getElementById("loginSuccess").innerHTML = ''
    document.getElementById("loginSuccess").style.display = 'none';
}

