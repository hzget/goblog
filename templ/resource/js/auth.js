        function signin(){
            username = document.getElementById("usr").value
            if (!validateUsername(username)) {
                displayDialog("Alert", "please input a valid username", "w3-red")
                return
            }
            password = document.getElementById("pwd").value
            creds = '{ "username": "' + username + '", "password": "' + password +'" }'
            const xhttp = new XMLHttpRequest();
            xhttp.onload = function() {
                result = this.responseText
                if (this.status == 200) {
                    location.href = "./"
                } else {
                    displayDialog("Alert", "failed to login:" + result, "w3-red")
                }
            }
            xhttp.open("POST", "./signin");
            xhttp.send(creds);
        }
        function logout(){
            username = document.getElementById("usr").value
            password = document.getElementById("pwd").value
            const xhttp = new XMLHttpRequest();
            xhttp.onload = function() {
                result = this.responseText
                if (this.status == 200) {
                    location.href = "./"
                } else {
                    displayDialog("Alert", "failed to logout", "w3-red")
                }
            }
            xhttp.open("POST", "./logout");
            xhttp.send();
        }

        function validateUsername(user) {
            let pattern = /^[0-9a-zA-Z]{3,10}$/g;
            return pattern.test(user);

        }

        function signup(){
            username = document.getElementById("usr").value
            if (!validateUsername(username)) {
                displayDialog("Alert", "please input a valid username", "w3-red")
                return
            }
            password = document.getElementById("pwd").value
            creds = '{ "username": "' + username + '", "password": "' + password +'" }'
            const xhttp = new XMLHttpRequest();
            xhttp.onload = function() {
                result = this.responseText
                if (this.status == 200) {
                    location.href = "./"
                } else {
                    displayDialog("Alert", "fail to register:" + result, "w3-red")
                }
            }
            xhttp.open("POST", "./signup");
            xhttp.send(creds);
        }

       function getCookie(cname) {
          let name = cname + "=";
          let decodedCookie = decodeURIComponent(document.cookie);
          let ca = decodedCookie.split(';');
          for(let i = 0; i < ca.length; i++) {
            let c = ca[i];
            while (c.charAt(0) == ' ') {
              c = c.substring(1);
            }
            if (c.indexOf(name) == 0) {
              return c.substring(name.length, c.length);
            }
          }
          return "";
        }

        function checkCookie() {
          let token = getCookie("session_token");
          let user = getCookie("user");
          if (token != "" && user != "" && token != null && user != null) {
              document.getElementById("login").setAttribute("hidden", true)
              document.getElementById("logout").removeAttribute("hidden")
              document.getElementById("welcome").innerHTML = "Welcome " + user + " !"
          } else {
              document.getElementById("login").removeAttribute("hidden")
              document.getElementById("logout").setAttribute("hidden", true)
              document.getElementById("welcome").innerHTML = "Login to read articles !"
          }
        }
        function openForm() {
          document.getElementById("auth-div").style.display = "block";
        }

        function closeForm() {
          document.getElementById("auth-div").style.display = "none";
        }
