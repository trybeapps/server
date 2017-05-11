$(function() {

	var userDropdownExists = false
			
	$('.user-label').on('mouseover', function() {
			
		$(this).children('img').attr('src', 'static/images/richard.jpeg').css('opacity', 1)
		$(this).children('svg').children('path').attr('stroke', '#DD4E4E')
			
	}).on('mouseleave', function() {
			
		if ( !userDropdownExists ) {
					
		  $(this).children('img').attr('src', 'static/images/richard_bw.jpeg').css('opacity', 0.55)
			$(this).children('svg').children('path').attr('stroke', '#BFBFBF')

		}
			
	}).on('click', function() {
				
	  var $userDropdown = $('.user-dropdown')
				
	  if ( $userDropdown.is(':visible') ) { hideUserDropDown() } else { showUserDropDown() }
			
	})

	function showUserDropDown() {

	  var $userDropdown = $('.user-dropdown')
		var $userLabel = $('.user-label')
				
		userDropdownExists = true
		$userLabel.children('svg').addClass('rotate')
		$userDropdown.addClass('anim-drop-down').show()

	}

	function hideUserDropDown() {

	  var $userDropdown = $('.user-dropdown')
		var $userLabel = $('.user-label')
				
		userDropdownExists = false
		$userLabel.children('svg').removeClass('rotate')
		$userDropdown.hide()
		$userLabel.trigger('mouseleave')
			
	}

  $('.book-list li').on('mouseover', function() {
    if ( !$(this).is(':first-child') ) $(this).addClass('book-jelly')
  }).on('mouseleave', function() {
    $(this).removeClass('book-jelly')
  })

  $('.search-label').click(function() {
    $('.book-list,.logo,.user-label,.user-dropdown').fadeOut(40)
    $('.search-label').fadeOut(40)
    $('.o-search-label').fadeIn(300).children('input[type="text"]').focus()
  })

  $('.o-search-label').on('click', 'svg', function() {
    $('.o-search-label').hide()
    $('.book-list,.logo,.user-label').show()
    $('.search-label').show()
  })

	$(document).on('click', function(e) {
	  if ($(e.target).closest('header .content').length === 0) hideUserDropDown()
	})

})