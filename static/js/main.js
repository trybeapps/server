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
    window.location.reload(false);
  })

  uploadBtn.on('success', function (file, res) {
    console.log(file)
    window.location.reload(false);
  })

  $('.user').click(function (e) {
    e.preventDefault()
    e.stopPropagation()
    $('.user-dropdown').toggle()
  })

  $(document).on('click', function () {
    $('.user-dropdown,.search-dropdown').hide()
  })

  $('#search').on('keyup', function () {
    if ($(this).val().length >= 3) {
      $('.search-dropdown').show()
      $.ajax({
        url: '/autocomplete',
        dataType: 'json',
        data: {
          term: $(this).val()
        },
        success: function (data) {
          console.log(data)
          $('.metadata ul').html('')
          if (!data[0].length) {
            $('.metadata-none').show()
          }
          else {
            $('.metadata-none').hide()
          }
          for (i in data[0]) {
            var title = data[0][i].title
            var author = data[0][i].author
            var url = data[0][i].url
            var cover = data[0][i].cover

            var html = '<li>'
                        + '<a href="' + url + '" class="sd-item">'
                          + '<img src="' + cover + '" width="60px" height="72px">'
                          + '<div class="sdi-info">'
                            + '<div class="sdii-title">' + title + '</div>'
                            + '<div class="sdii-author">' + author + '</div>'
                          + '</div>'
                        + '</a>'
                        + '</li>'

            $('.metadata ul').append(html)
          }

          $('.content ul').html('')
          if (!data[1].length) {
            $('.content-none').show()
          }
          else {
            $('.content-none').hide()
          }
          for (i in data[1]) {
            var title = data[1][i].title
            var author = data[1][i].author
            var url = data[1][i].url
            var cover = data[1][i].cover
            var page = data[1][i].page
            var content = data[1][i].data

            var html = '<li>'
                        + '<a href="' + url + '#page=' + page + '" class="sd-item">'
                          + '<img src="' + cover + '" width="60px" height="72px">'
                          + '<div class="sdi-info">'
                            + '<div class="sdii-title">' + title + '</div>'
                            + '<div class="sdii-author">' + author + '</div>'
                            + '<div class="sdii-content">' + content + '</div>'
                          + '</div>'
                        + '</a>'
                        + '</li>'

            $('.content ul').append(html)
          }
        }
      })
    }
  })
})
