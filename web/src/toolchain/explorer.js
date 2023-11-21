"use strict"
import '../../public/assets/js/jquery-1.11.0.min.js'

function getQueryVariable(variable) {
    var query = window.location.search.substring(1);
    var vars = query.split("&");
    for (var i=0;i<vars.length;i++) {
            var pair = vars[i].split("=");
            if(pair[0] == variable){return pair[1];}
    }
    return '';
}

function FilesRender(params){
    let beforeOpen = params['beforeOpen']
    let redirect = params['redirect']
    let readonly = params['readonly']
    let from = params['from']
    let resp = params['resp']

    let filemanager = $('.filemanager')
    let breadcrumbs = $('.breadcrumbs')
    let fileList = filemanager.find('.data')

    // // Start by fetching the file data from scan route with an AJAX request
    // initFunc(scanUrl, function(resp) {
        let respdata = resp.data;
        let response = [respdata]
        let currentPath = ''
        let breadcrumbsUrls = []
        let folders = []
        let files = []
        // let data = JSON.parse(respdata)
        // response = [data]
        let data = respdata
        wrapGoto(redirect)
        // goto("encryption/output-1642042423396842384")

        // /* This event listener monitors changes on the URL. We use it to
        // *  capture back/forward navigation in the browser.
        // */
        // $(window).on('hashchange', function() {
        //     //console.log(window.location.hash)
        //     if (window.location.hash !== '#/') {
        //         window.location.hash = '#/';
        //         window.location.href = '/?page=res'
        //     }

        //     /* We are triggering the event. This will execute 
        //     *  this function on page load, so that we show the correct folder:
        //     */
        // }).trigger('hashchange')

        // Hiding and showing the search box
        filemanager.find('.search').click(function() {
            let search = $(this)
            search.find('span').hide()
            search.find('input[type=search]').show().focus()
        })


        /* Listening for keyboard input on the search field.
        *  We are using the "input" event which detects cut and paste
        *  in addition to keyboard input.
        */
        filemanager.find('input').on('input', function(e) {
            folders = []
            files = []

            let value = this.value.trim()

            if (value.length) {
                filemanager.addClass('searching')

                // Update the hash on every key stroke
                // window.location.hash = 'search=' + value.trim()
                wrapGoto('search=' + value.trim())
            }
            else {
                filemanager.removeClass('searching')
                // window.location.hash = encodeURIComponent(currentPath)
                wrapGoto(encodeURIComponent(currentPath))
            }
        }).on('keyup', function(e) {
            // Clicking 'ESC' button triggers focusout and cancels the search
            let search = $(this)

            if(e.keyCode == 27) {
                search.trigger('focusout')
            }

        }).focusout(function(e) {
            // Cancel the search
            let search = $(this)

            if(!search.val().trim().length) {
                // window.location.hash = encodeURIComponent(currentPath)
                search.hide()
                search.parent().find('span').show()
                wrapGoto(encodeURIComponent(currentPath))
            }
        })


        // Clicking on folders
        fileList.on('click', 'li.folders', function(e) {
            e.preventDefault()

            let nextDir = $(this).find('a.folders').attr('href')
            if (beforeOpen(nextDir) === false) {
                return
            }

            if(filemanager.hasClass('searching')) {
                // Building the breadcrumbs
                breadcrumbsUrls = generateBreadcrumbs(nextDir)

                filemanager.removeClass('searching')
                filemanager.find('input[type=search]').val('').hide()
                filemanager.find('span').show()
            }
            else {
                breadcrumbsUrls.push(nextDir)
            }
            // window.location.hash = encodeURIComponent(nextDir)
            currentPath = nextDir
            wrapGoto(encodeURIComponent(nextDir))
        })


        // Clicking on breadcrumbs
        breadcrumbs.on('click', 'a', function(e){
            e.preventDefault()

            let index = breadcrumbs.find('a').index($(this))
            let nextDir = breadcrumbsUrls[index]
            if (beforeOpen(nextDir) === false) {
                return
            }
            if (index < 0) return;
            breadcrumbsUrls.length = Number(index)

            // window.location.hash = encodeURIComponent(nextDir)
            wrapGoto(encodeURIComponent(nextDir))
        })

        function wrapGoto(hash) {
            let redirectPath = hash;
            if (redirectPath.indexOf('#') === 0) {
                redirectPath = redirectPath.split('#')[1];
            }
            if (redirectPath !== '' || redirectPath.indexOf('/') > 0) redirectPath = '/' + redirectPath
            currentPath = redirectPath;
           
            //console.log('redirectPath1:'+redirectPath)
            //console.log('window.location.hash:'+window.location.hash)
            goto(redirectPath)
        }

        // Navigates to the given hash (path)
        function goto(hash) {
            hash = decodeURIComponent(hash).slice(1).split('=')

            if (hash.length) {
                let rendered = ''

                // if hash has search in it
                if (hash[0] === 'search') {
                    filemanager.addClass('searching')
                    rendered = searchData(response, hash[1].toLowerCase())

                    if (rendered.length) {
                        currentPath = hash[0]
                        render(rendered)
                    }
                    else {
                        render(rendered)
                    }
                }

                // if hash is some path
                else if (hash[0].trim().length) {
                    rendered = searchByPath(hash[0])
                    if (rendered.length) {
                        currentPath = hash[0]
                        breadcrumbsUrls = generateBreadcrumbs(hash[0])
                        render(rendered)
                    }
                    else {
                        currentPath = hash[0]
                        breadcrumbsUrls = generateBreadcrumbs(hash[0])
                        render(rendered)
                    }
                }

                // if there is no hash
                else {
                    currentPath = data.path
                    breadcrumbsUrls.push(data.path)
                    render(searchByPath(data.path))
                }
            }
        }

        // Splits a file path and turns it into clickable breadcrumbs
        function generateBreadcrumbs(nextDir){
            let path = nextDir.split('/').slice(0)
            for (let i=1; i<path.length; i++){
                path[i] = path[i-1]+ '/' +path[i]
            }
            return path
        }


        // Locates a file by path
        function searchByPath(dir) {
            let path = dir.split('/')
            let demo = response
            let flag = 0

            //console.log('dir:' + dir)
            //console.log('path:' + path)
            //console.log('demo:' + JSON.stringify(demo))
            for (let i=0; i<path.length; i++) {
                for (let j=0; j<demo.length; j++) {
                    //console.log(demo[j].name)
                    //console.log(path[i])
                    if (demo[j].name === path[i]) {
                        flag = 1
                        demo = demo[j].items
                        break
                    }
                }
            }

            demo = flag ? demo : []
            return demo
        }

        // Recursively search through the file tree
        function searchData(data, searchTerms) {
            data.forEach(function(d) {
                if (d.type === 'folder') {
                    searchData(d.items,searchTerms)

                    if (d.name.toLowerCase().match(searchTerms)) {
                        folders.push(d)
                    }
                }
                else if (d.type === 'file') {
                    if (d.name.toLowerCase().match(searchTerms)) {
                        files.push(d)
                    }
                }
            })
            return {folders: folders, files: files}
        }

        // Render the HTML for the file manager
        function render(data) {
            let scannedFolders = []
            let scannedFiles = []

            if (Array.isArray(data)) {
                data.forEach(function(d) {
                    if (d.type === 'folder') {
                        scannedFolders.push(d)
                    }
                    else if (d.type === 'file') {
                        scannedFiles.push(d)
                    }
                })
            }
            else if (typeof data === 'object') {
                scannedFolders = data.folders
                scannedFiles = data.files
            }

            // Empty the old result and make the new one
            fileList.empty().hide()

            if (!scannedFolders.length && !scannedFiles.length) {
                filemanager.find('.nothingfound').show()
            }
            else {
                filemanager.find('.nothingfound').hide()
            }

            if (scannedFolders.length) {
                scannedFolders.forEach(function(f) {
                    let itemsLength = f.items.length
                    let name = escapeHTML(f.name)
                    let icon = '<span class="icon folder"></span>'

                    if (itemsLength) {
                        icon = '<span class="icon folder full"></span>'
                    }
                    if (itemsLength == 1) {
                        itemsLength += ' item'
                    }
                    else if (itemsLength > 1) {
                        itemsLength += ' items'
                    }
                    else {
                        itemsLength = 'Empty'
                    }

                    let dirUrl = f.path;
                    // let dirUrl = '/api/file/scan?path=' + f.path;
                    let eleLink = 'href="' + dirUrl + '" title="'+ f.path +'" class="folders">'
                    let eleDesc = '<span class="name">' + name + '</span> <span class="details">' + itemsLength + '</span>'
                    let folder = $('<li class="folders"><a ' + eleLink + icon + eleDesc + '</a></li>')
                    folder.appendTo(fileList)
                })
            }

            if (scannedFiles.length) {
                scannedFiles.forEach(function(f) {
                    let fileSize = bytesToSize(f.size)
                    let name = escapeHTML(f.name)
                    let fileType = name.split('.')
                    let icon = '<span class="icon file"></span>'

                    fileType = fileType.length > 1 ? fileType[fileType.length-1] : ''

                    icon = '<span class="icon file f-' + fileType + '">' + fileType + '</span>'

                    let file = null;
                    // let fileUrl = '/api/file/preview?page=view&readonly=false&filepath=' + f.path + '&from=' + fromApp;
                    let fileUrl = '/api/file/preview?page=view&readonly=' + readonly + '&filepath=' + f.path + '&from=' + from;
                    let isImg = name.indexOf('.png') > 0 || name.indexOf('.jpg') > 0 || name.indexOf('.jpeg') > 0 || name.indexOf('.gif') > 0 || name.indexOf('.bmp') > 0;
                    if (isImg) file = $('<li class="files"><a data-fancybox="gallery" href="'+ fileUrl+'" title="'+ f.path +'" class="files">'+icon+'<span class="name">'+ name +'</span> <span class="details">'+fileSize+'</span></a></li>');
                    else file = $('<li class="files"><a href="'+ fileUrl+'" title="'+ f.path +'" class="files">'+icon+'<span class="name">'+ name +'</span> <span class="details">'+fileSize+'</span></a></li>')
                    file.appendTo(fileList)
                })
            }


            // Generate the breadcrumbs
            let url = ''

            if (filemanager.hasClass('searching')) {
                url = '<span>Search results: </span>'
                fileList.removeClass('animated')
            }
            else {
                fileList.addClass('animated')
                breadcrumbsUrls.forEach(function (u, i) {
                    let name = u.split('/')

                    if (i !== breadcrumbsUrls.length - 1) {
                        url += '<a href="'+u+'"><span class="folderName">' + name[name.length-1] + '</span></a> <span class="arrow">→</span> '
                    }
                    else {
                        url += '<span class="folderName">' + name[name.length-1] + '</span>'
                    }
                })
            }

            breadcrumbs.text('').append(url)


            // Show the generated elements
            fileList.animate({'display':'inline-block'})
        }


        // This function escapes special html characters in names
        function escapeHTML(text) {
            return text.replace(/\&/g,'&amp;').replace(/\</g,'&lt;').replace(/\>/g,'&gt;')
        }


        // Convert file sizes from bytes to human readable units
        function bytesToSize(bytes) {
            let sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
            if (bytes == 0) return '0 Bytes'
            let i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)))
            return Math.round(bytes / Math.pow(1024, i), 2) + ' ' + sizes[i]
        }
    // })
}
export {FilesRender}
