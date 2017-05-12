$(function() {

	var userDropdownExists = false
			
	$('.user-label').on('mouseover', function() {
			
		$(this).children('img').attr('src', '/static/images/richard.jpeg').css('opacity', 1)
		$(this).children('svg').children('path').attr('stroke', '#DD4E4E')
			
	}).on('mouseleave', function() {
			
		if ( !userDropdownExists ) {
					
		  $(this).children('img').attr('src', '/static/images/richard_bw.jpeg').css('opacity', 0.55)
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

  $('.book-list li,.collection-list li').on('mouseover', function() {
    if ( !$(this).hasClass('anb-item') && !$(this).hasClass('cd-item') && !$(this).hasClass('anc-item') ) $(this).addClass('book-jelly')
  }).on('mouseleave', function() {
    $(this).removeClass('book-jelly')
  })

  $('.search-label').click(function() {
    showSearch()
  })

  $('.o-search-label').on('click', 'svg', function(e) {
    history.back()
  })

  $('.collection-label').on('click', function(e) {
    e.preventDefault()
    if (window.location.href.split('/').pop() != 'collections') showCollection()
  })

  $('.collection-container a').on('click', function(e) {
    e.preventDefault()
    if ( !$(this).closest('li').hasClass('anc-item') ) {
      var collectionTitle = $(this).attr('href').split('/').pop()
      setHistory('collection-detail', '/collections/' + collectionTitle)
      $('.book-container').addClass('collection-detail')
      $('.collection-container,.anb-item').hide()
      $('.book-container,.cd-item').show()
    }
  })

  function showSearch() {
    setHistory('search', '/search')
    $('.book-container,.logo,.header-link,.user-label,.user-dropdown,.collection-container').fadeOut(40)
    $('.search-label').fadeOut(40)
    $('.search-container').show()
    $('.o-search-label').fadeIn(300).children('input[type="text"]').focus()
  }

  function showHome() {
    setHistory('home', '/')
    $('.o-search-label,.search-container,.collection-container,.collection-detail-container').hide()
    $('.book-container,.logo,.header-link,.user-label').show()
    $('.collection-label').css('border-bottom', 'none')
    $('.search-label').show()
  }

  function showCollection(e) {
    setHistory('collections', '/collections')
    $('.o-search-label,.search-container,.book-container').fadeOut(40)
    $('.collection-label').css('border-bottom', '3px solid #DD4E4E')
    $('.collection-container,.search-label,.logo,.header-link,.user-label').show()
  }

  function setHistory(page, url) {
    history.pushState({
      page: page
    },null,url)
  }

  setHistory('home', '/')

  window.onpopstate = function (e) {
    if (e.state.page == 'home') {
      showHome()
    } else if (e.state.page == 'collections') {
      showCollection()
    }
  }

	$(document).on('click', function(e) {
	  if ($(e.target).closest('header .content').length === 0) hideUserDropDown()
	})

})