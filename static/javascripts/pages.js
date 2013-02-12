$(window).ready(function(){

	// Make list sortable
	$(function() {
		$("#sortable").sortable({
			update: function(event, ui) {
				var order = new Array();
				$("#sortable > li").each(function() {
					order.push($(this).attr("id"));
				});
				$.ajax({
					type: "POST",
					url: "/manage/pages/sort",
					data: { 'order': order.join(",") }
				})
			}
		});
		$("#sortable").disableSelection();
	});

});
