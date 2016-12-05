// Disabling autoDiscover, otherwise Dropzone will try to attach twice.
Dropzone.autoDiscover = false

$(function () {
  var myDropzone = new Dropzone('body', {
    url: '/book-upload',
    paramName: 'file',
    acceptedFiles: 'application/pdf',
    uploadMultiple: true,
    clickable: false
  })

  myDropzone.on('dragover', function () {
    $('.dz-message').fadeIn()
    return setTimeout(function () {
      $('.dz-message').fadeOut()
    }, 2000)
  })

  var uploadBtn = new Dropzone('.upload-btn', {
    url: '/book-upload',
    paramName: 'file',
    acceptedFiles: 'application/pdf',
    uploadMultiple: true
  })

  $('.upload-btn').click(function (e) {
    $('.dz-hidden-input').click()
  })

  myDropzone.on('success', function(file,res) {
    console.log(file)
    alert(file.name + ' uploaded successfully')
  })

  uploadBtn.on('success', function(file,res) {
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
})
