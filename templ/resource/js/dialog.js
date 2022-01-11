
        function getdoc() {
            return window.parent.document;
        }
        function resetDialog() {
            doc = getdoc()
            header = doc.getElementById("dialogheader")
            footer = doc.getElementById("dialogfooter")
            header.removeAttribute("class");
            footer.removeAttribute("class");
            header.setAttribute("class", "w3-container");
            footer.setAttribute("class", "w3-container");
            doc.getElementById("dialogheaderinfo").innerHTML = ""
            doc.getElementById("dialoginfo").innerHTML = ""
            doc.getElementById('dialogbox').style.display='none'
        }

        function setDialogColor(color) {
            doc = getdoc()
            header = doc.getElementById("dialogheader")
            footer = doc.getElementById("dialogfooter")
            header.classList.add(color)
            footer.classList.add(color)
        }

        function confirmResult(){
            resetDialog()
        }

        function displayDialog(header, info, color="w3-blue"){
            doc = getdoc()
            setDialogColor(color)
            doc.getElementById("dialogheaderinfo").innerHTML = header
            doc.getElementById("dialoginfo").innerHTML = info
            doc.getElementById('dialogbox').style.display='block'
        }

