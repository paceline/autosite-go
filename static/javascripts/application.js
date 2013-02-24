$(window).ready(function(){
  
	// Do menu highlighting
	$("#menu > a").each(function() {
		var path = window.location.pathname.split("/");
		if (path.length > 3) {
			path = path.slice(0,3);
		}
		if ($(this).attr("href") == path.join("/")) {
			$(this).addClass("selected");
		}
		else {
			$(this).removeClass("selected");
		}
	});
	
	// Hide Notice after 5 seconds
	if ($('#notice')){
		setTimeout("$('#notice').hide()", 5000);
    }
    
    // Ajax Call detector
    $('a[data-remote="true"]').click(function() {
		event.preventDefault();
		var method = ($(this).attr('data-method') == null ? 'get' : $(this).attr('data-method'));
		var path = $(this).attr('href').split('/');
		switch (method) {
			case 'delete':
				$.ajax({
					type: method,
					url: $(this).attr('href'),
					complete: function() {
						$('<div id="notice"><p>"' + toTitle(path[path.length-1]) + '" has been deleted</p></div>').insertAfter('#menu');
						setTimeout("$('#notice').hide()", 5000);
					}
				})
				break;
		}
	});
	
	// Helper for upcasing just the first char
	function toTitle(word) {
		return word.charAt(0).toUpperCase() + word.slice(1)
	}
    
});
