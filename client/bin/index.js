const issuer = "https://{DOMAIN}.okta.com/oauth2/default";


const windowFeatures = "menubar=yes,location=yes,resizable=yes,scrollbars=yes,status=no,width=570px,height=720px";
// const windowFeatures = "menubar=yes,location=yes,resizable=yes,scrollbars=yes,status=yes"
var windowObjectReference;
   

window.addEventListener("message", (event) => {
    // Do we trust the sender of this message?
    // if (event.origin !== "http://example.com:8080")
    //   return;
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
    document.getElementById("loginError").innerHTML = ''
    document.getElementById("loginError").style.display = 'none';
    document.getElementById("loginSuccess").innerHTML = ''
    document.getElementById("loginSuccess").style.display = 'none';
    
    const authClient = new OktaAuth({
        issuer: issuer
    });

    authClient.signInWithCredentials({
    username: document.getElementById("userName").value,
    password: document.getElementById("password").value
    })
    .then(function(transaction) {
    if (transaction.status === 'SUCCESS') {
        document.getElementById("userName").value = '';
        document.getElementById("password").value = '';
        // calling /authorize direct from the client means server can't match the state
        // const directAuthorizeUrl = issuer + '/v1/authorize?' +
        //                     'client_id=' + clientId + '&response_type=code&response_mode=query&' +
        //                     'scope=openid%20profile%20email&' +
        //                     'redirect_uri=' + redirectURI + '&state=foreverInTheSameState&' +
        //                     'nonce=85582b03-f422-4742-b515-eedefe373ae2&sessionToken=' + 
        //                     transaction.sessionToken;
        const directAuthorizeUrl = `http://localhost:8082/authorize?session_token=${transaction.sessionToken}`
        windowObjectReference = window.open(directAuthorizeUrl, '_blank', windowFeatures);
    } else {
        throw 'We cannot handle the ' + transaction.status + ' status';
    }
    })
    .catch(function(err) {
        document.getElementById("loginError").innerHTML = err;
        document.getElementById("loginError").style.display = 'block';
    });
}

