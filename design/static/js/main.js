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
})