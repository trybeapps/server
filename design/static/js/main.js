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
	var crcbHeight = parseInt($('.crcb-book-list a:first-child img')[0].getBoundingClientRect().height)
	$('.crcb-book-list').css('height', crcbHeight + 'px')
	$('.crcb-book-list a').each(function() {
		var thisHeight = parseInt($(this).children('img')[0].getBoundingClientRect().height)
		if ( thisHeight > crcbHeight ) $('.crcb-book-list').css('height', thisHeight + 'px')
		$(this).css('left', crcbCounter + 'px')
		crcbCounter = crcbCounter + parseInt($(this).children('img').width()) + 30
	})

	function getScrollBarWidth () {
  		var inner = document.createElement('p');
  		inner.style.width = "100%";
  		inner.style.height = "200px";

  		var outer = document.createElement('div');
  		outer.style.position = "absolute";
  		outer.style.top = "0px";
  		outer.style.left = "0px";
  		outer.style.visibility = "hidden";
  		outer.style.width = "200px";
  		outer.style.height = "150px";
  		outer.style.overflow = "hidden";
  		outer.appendChild (inner);

  		document.body.appendChild (outer);
  		var w1 = inner.offsetWidth;
  		outer.style.overflow = 'scroll';
  		var w2 = inner.offsetWidth;
  		if (w1 == w2) w2 = outer.clientWidth;

  		document.body.removeChild (outer);

  		return (w1 - w2);
	};

	var windowWidth = $(window).width() + parseInt(getScrollBarWidth())

	var booksLength = 6
	var crcbImgWidthFULL = 0
	var crcbImgWidthPartial = 0
	if ( windowWidth > 1300 ) {
		booksLength = 6
		crcbImgWidthFULL = ( parseInt( $('.crcb-book-list a img').width() ) * 6 ) + 180
		crcbImgWidthPartial = ( parseInt($('.crcb-book-list a img').width()) * 5 ) + 150 
	} else if ( windowWidth <= 1300 && windowWidth >= 900 ) {
		booksLength = 4
		crcbImgWidthFULL = ( parseInt( $('.crcb-book-list a img').width() ) * 4 ) + 120
		crcbImgWidthPartial = ( parseInt($('.crcb-book-list a img').width()) * 3 ) + 90
	} else if ( windowWidth <= 899 && windowWidth >= 470 ) {
		booksLength = 2
		crcbImgWidthFULL = ( parseInt( $('.crcb-book-list a img').width() ) * 2 ) + 60
		crcbImgWidthPartial = ( parseInt($('.crcb-book-list a img').width()) * 1 ) + 30
	} else if ( windowWidth <= 469 && windowWidth >= 300 ) {
		booksLength = 1
		crcbImgWidthFULL = ( parseInt( $('.crcb-book-list a img').width() ) * 1 ) + 30
		crcbImgWidthPartial = 0
	}

	$('.crcb-arrow .right').click(function() {

		if ($('.crcb-book-list a').length > booksLength) {

			$('.crcb-arrow .left').removeClass('none')

			if ( parseInt($('.crcb-book-list a:last-child').css('left').split('px')[0]) != crcbImgWidthPartial ) {
				$('.crcb-book-list a').each(function() {
					var left = parseInt($(this).css('left').split('px')[0]) - ( parseInt($(this).children('img').width()) + 30 )
					$(this).css('left', left + 'px')
				})

				if ( parseInt($('.crcb-book-list a:last-child').css('left').split('px')[0]) ==  crcbImgWidthFULL ) {
					$('.crcb-arrow .right').addClass('none')
				} else {
					$('.crcb-arrow .right').removeClass('none')
				}

			}

		}

	})

	$('.crcb-arrow .left').click(function() {

		if ($('.crcb-book-list a').length > booksLength) {
			
			$('.crcb-arrow .right').removeClass('none')

			if ( parseInt($('.crcb-book-list a:first-child').css('left').split('px')[0]) != 0 ) {

				$('.crcb-book-list a').each(function() {
					var left = parseInt($(this).css('left').split('px')[0]) + ( parseInt($(this).children('img').width()) + 30 )
					$(this).css('left', left + 'px')
				})
			
				if ( parseInt($('.crcb-book-list a:first-child').css('left').split('px')[0]) == -Math.abs(parseInt($('.crcb-book-list a img').width())) - 30 ) {
					$('.crcb-arrow .left').addClass('none')
				} else {
					$('.crcb-arrow .left').removeClass('none')
				}
			
			}

		}

	})
})