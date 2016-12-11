// Disable auto discover for all elements:
Dropzone.autoDiscover = false;

$(function () {
  var myDropzone = new Dropzone('body', {
    url: '/book-upload',
    paramName: 'file',
    acceptedFiles: 'application/pdf',
    uploadMultiple: true,
    clickable: false
  })

  myDropzone.on('dragover', function () {
    $('.page-container').addClass('drag-over')
  })

  myDropzone.on('dragleave', function () {
      $('.page-container').removeClass('drag-over')
  })

  myDropzone.on('drop', function () {
      $('.page-container').removeClass('drag-over')
  })

  var uploadBtn = new Dropzone('.upload-books', {
    url: '/book-upload',
    paramName: 'file',
    acceptedFiles: 'application/pdf',
    uploadMultiple: true
  })

  myDropzone.on('success', function (file, res) {
    console.log(file)
    alert(file.name + ' uploaded successfully')
  })

  uploadBtn.on('success', function (file, res) {
    console.log(file)
    alert(file.name + ' uploaded successfully')
  })

  $('.search-box').on('focusin', '#search', function () {
    $('.search-icon').children('g').css('stroke', '#f34a53')
  })

  $('.search-box').on('focusout', '#search', function () {
    $('.search-icon').children('g').css('stroke', '#8B8B8B')
  })

  $('.user-item').click(function (e) {
    e.preventDefault()
    e.stopPropagation()
    $('.user-dropdown').toggle()
  })

  $(document).on('click', function () {
    $('.user-dropdown').hide()
  })

  $('#search').on('keyup', function () {
    if ($(this).val().length >= 3) {
      $('.autocomplete').show()
      $.ajax({
        url: '/autocomplete',
        dataType: 'json',
        data: {
          term: $(this).val()
        },
        success: function (data) {
          console.log(data)
          $('.atc-metadata').html('')
          for (i in data[0]) {
            var title = data[0][i].title
            var author = data[0][i].author
            var url = data[0][i].url
            var cover = data[0][i].cover

            var html = '<li>'
                          + '<a href="' + url + '">'
                            + '<img src="' + cover + '" width="60px" height="70px">'
                            + '<div class="atc-b-info">'
                              + '<div class="atc-b-title">' + title + '</div>'
                              + '<div class="atc-b-author">' + author + '</div>'
                            + '</div>'
                          + '</a>'
                        + '</li>'

            $('.atc-metadata').append(html)
          }

          $('.atc-content').html('')
          for (i in data[1]) {
            var title = data[1][i].title
            var author = data[1][i].author
            var url = data[1][i].url
            var cover = data[1][i].cover
            var page = data[1][i].page
            var content = data[1][i].data

            var html = '<li>'
                          + '<a href="' + url + '#page=' + page + '">'
                            + '<img src="' + cover + '" width="60px" height="70px">'
                            + '<div class="atc-b-full-text">'
                              + '<div class="atc-b-title">' + title + '</div>'
                              + '<div class="atc-b-author">' + author + '</div>'
                              + '<svg width="14" height="12" viewBox="0 0 14 12" xmlns="http://www.w3.org/2000/svg">'
                                + '<title>Shape</title>'
                                + '<path d="M2.646 2.679C1.609 3.975 1.459 5.28 1.764 6.195c1.151-.914 2.751-.723 3.726.188.984.92 1.069 2.536.44 3.582a2.88 2.88 0 0 1-2.49 1.411C1.082 11.376 0 9.298 0 6.966 0 5.454.386 4.098 1.157 2.9 1.93 1.701 3.094.735 4.652 0l.419.816c-.94.397-1.75 1.018-2.425 1.863zm7.672 0C9.281 3.975 9.132 5.28 9.436 6.195c.515-.397 1.073-.595 1.676-.595C12.697 5.6 14 6.656 14 8.488c0 1.685-1.293 2.888-2.888 2.888-2.356 0-3.44-2.078-3.44-4.41 0-1.513.386-2.869 1.158-4.067C9.602 1.701 10.766.735 12.324 0l.42.816c-.941.397-1.75 1.018-2.426 1.863z" fill="#0766B4" fill-rule="evenodd"/>'
                              + '</svg>'
                              + '<p>' + content + '</p>'
                            + '</div>'
                          + '</a>'
                        + '</li>'

            $('.atc-content').append(html)
          }
        }
      })
    }
  })
})
