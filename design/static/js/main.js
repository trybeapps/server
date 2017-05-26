$(function() {
	$('.hn-user-nav').click(function(e) {
		e.preventDefault()
		if ( !$('.user-dropdown').is(':visible') ) {
			$('.user-dropdown').show()
		} else {
			$('.user-dropdown').hide()
		}
	})

	$(document).click(function(e) {
		if ( $(e.target).closest('.user-dropdown').length === 0 && $(e.target).closest('.hn-user-nav').length === 0 ) {
			$('.user-dropdown').hide()
		}
	})

	if ($('.crcb-book-list a').length <= 6) {
		$('.crcb-arrow div').addClass('none')
	}

	var crcbCounter = 0
	$('.crcb-book-list a').each(function() {
		$(this).css('left', crcbCounter + 'px')
		crcbCounter = crcbCounter + 235
	})

	$('.crcb-arrow .right').click(function() {
		if ($('.crcb-book-list a').length > 6) {
			
			$('.crcb-arrow .left').removeClass('none')

			if ( parseInt($('.crcb-book-list a:last-child').css('left').split('px')[0]) != 1175 ) {
				$('.crcb-book-list a').each(function() {
					var left = parseInt($(this).css('left').split('px')[0]) - 235
					$(this).css('left', left + 'px')
				})

				if ( parseInt($('.crcb-book-list a:last-child').css('left').split('px')[0]) == 1410 ) {
					$('.crcb-arrow .right').addClass('none')
				} else {
					$('.crcb-arrow .right').removeClass('none')
				}

			}

		}

	})

	$('.crcb-arrow .left').click(function() {
		if ($('.crcb-book-list a').length > 6) {
			
			$('.crcb-arrow .right').removeClass('none')

			if ( parseInt($('.crcb-book-list a:first-child').css('left').split('px')[0]) != 0 ) {

				$('.crcb-book-list a').each(function() {
					var left = parseInt($(this).css('left').split('px')[0]) + 235
					$(this).css('left', left + 'px')
				})
			
				if ( parseInt($('.crcb-book-list a:first-child').css('left').split('px')[0]) == -235 ) {
					$('.crcb-arrow .left').addClass('none')
				} else {
					$('.crcb-arrow .left').removeClass('none')
				}
			
			}

		}
	})
})