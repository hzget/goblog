        function signin(){
            username = document.getElementById("usr").value
            password = document.getElementById("pwd").value
            creds = '{ "username": "' + username + '", "password": "' + password +'" }'
            const xhttp = new XMLHttpRequest();
            xhttp.onload = function() {
                result = this.responseText
                if (this.status == 200) {
                    location.href = "./"
                } else {
                    alert("failed to login")
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
                    alert("failed to logout")
                }
            }
            xhttp.open("POST", "./logout");
            xhttp.send();
        }
        function signup(){
            username = document.getElementById("usr").value
            password = document.getElementById("pwd").value
            creds = '{ "username": "' + username + '", "password": "' + password +'" }'
            const xhttp = new XMLHttpRequest();
            xhttp.onload = function() {
                result = this.responseText
                if (this.status == 200) {
                    location.href = "./"
                } else {
                    alert("fail to register:" + result)
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
          if (token != "" && user != "") {
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
