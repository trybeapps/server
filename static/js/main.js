$(function() {

	// Hide Header on on scroll down
	var didScroll;
	var lastScrollTop = 0;
	var delta = 5;
	var navbarHeight = $('header').outerHeight();

	$(window).scroll(function(event){
    	didScroll = true;
	});

	setInterval(function() {
    	if (didScroll) {
        	hasScrolled();
        	didScroll = false;
    	}
	}, 250);

	function hasScrolled() {
    	var st = $(this).scrollTop();
    
    	// Make sure they scroll more than delta
    	if(Math.abs(lastScrollTop - st) <= delta)
        	return;
    
    	// If they scrolled down and are past the navbar, add class .nav-up.
    	// This is necessary so you never see what is "behind" the navbar.
    	if (st > lastScrollTop && st > navbarHeight){
        	// Scroll Down
        	$('header').removeClass('nav-down').addClass('nav-up');
    	} else {
        	// Scroll Up
        	if(st + $(window).height() < $(document).height()) {
            	$('header').removeClass('nav-up').addClass('nav-down');
        	}
    	}
    
    	lastScrollTop = st;
	}

	$('.menu-icon').click(function() {
		$('.header-nav-small').show()
		$('body').css('overflow-y', 'hidden')
	})

	$('.hns-close').click(function() {
		$('.header-nav-small').hide()
		$('body').css('overflow-y', 'auto')
	})

	if ($('.crcb-book-list a').length <= 6) {
		$('.crcb-arrow div').addClass('none')
	}

	var crcbCounter = 0
	var crcbHeight = 0
	if ($('.crcb-book-list a').length > 0) {
		crcbHeight = parseInt($('.crcb-book-list a:first-child img')[0].getBoundingClientRect().height)
	}
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
	} else if ( windowWidth <= 899 && windowWidth >= 520 ) {
		booksLength = 2
		crcbImgWidthFULL = ( parseInt( $('.crcb-book-list a img').width() ) * 2 ) + 60
		crcbImgWidthPartial = ( parseInt($('.crcb-book-list a img').width()) * 1 ) + 30
	} else if ( windowWidth <= 519 && windowWidth >= 300 ) {
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

	$('.hn-book-nav').click(function(e) {
		e.preventDefault()
		$('.upload-books').click()
	})

	$('.upload-books').change(function(e) {
		e.preventDefault()
		// var ext = $(this).val().split('.').pop().toLowerCase();
		// alert(ext)
		var files = $(this).get(0).files
		for (var i = 0; i < files.length; i++) {
			if (files[i].type == 'application/pdf') {
				console.log(files[i])
			}
			else {
				alert(files[i].name + '(Wrong file format. Only PDFs are supported.)')
				continue
			}
		}
		$('.upload-books-form').submit()
	})

	$('.upload-books-form').submit(function(e) {
		e.preventDefault()
		var formData = new FormData($(this)[0])
		console.log(formData)
		$.ajax({
			url: '/upload',
			type: 'POST',
			data: formData,
			contentType: false,
			cache: false,
			processData:false,
            success: function (data) {
                alert(data)
                window.location.href = "/"
            }
		})
	})

	$('.bc-pagination .none').click(function(e) {
		e.preventDefault()
	})

	$('.search-box').on('keyup', function(e) {
		if ($(this).val().length >= 3) {
			$('.search-dropdown').show()
			$.ajax({
				url: '/autocomplete',
				dataType: 'json',
				data: {
					term: $(this).val()
				},
				success: function(data) {
					console.log(data['book_detail'][0]['highlight']['attachment.content'][0])
				}
			})
		}
	})

	$(document).click(function(e) {
		if ( $(e.target).closest('.search-dropdown').length == 0 && $(e.target).closest('.search-box').length == 0 ) {
			$('.search-dropdown').hide()
		}
	})
})