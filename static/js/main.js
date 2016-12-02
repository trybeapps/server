$(function () {
  $('.search-box').on('focusin', '#search', function () {
    $('.search-icon').children('g').css('stroke', '#0766B4')
  })

  $('.search-box').on('focusout', '#search', function () {
    $('.search-icon').children('g').css('stroke', '#8B8B8B')
  })

  $('.user-item').click(function (e) {
    e.preventDefault();
    e.stopPropagation();
    $('.user-dropdown').toggle();
  })

  $(document).on('click', function () {
    $('.user-dropdown').hide();
  })
})
