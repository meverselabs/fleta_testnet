(function($) {
	"use strict";
	
	$(document).ready(function() {
		/*  [ Page loader ]
		- - - - - - - - - - - - - - - - - - - - */
		// $( 'body' ).addClass( 'loaded' );
		// setTimeout(function () {
		// 	$('#page-loader').fadeOut();
		// }, 500);
        
		/*  [ Main Menu ]
        - - - - - - - - - - - - - - - - - - - - */
		$( '.sub-menu' ).each(function() {
            $( this ).parent().addClass( 'has-child' ).find( '> a' ).append( '<span class="arrow"><i class="fa fa-angle-right"></i></span>' );
        });
        $( '.main-menu .arrow' ).on( 'click', function(e) {
            e.preventDefault();
            $( this ).parents( 'li' ).find( '> .sub-menu' ).slideToggle( 'fast' );
        });
        $( '.mobile-btn' ).on( 'click', function() {
            $( this ).parents( '.main-menu' ).toggleClass('open');
            $( 'body' ).toggleClass( 'menu-open' );
        });

        $( 'html' ).on( 'click', function(e) {
            if( $( e.target ).closest( '.main-menu.open' ).length == 0 ) {
                $( '.main-menu' ).removeClass( 'open' );
                $( 'body' ).removeClass( 'menu-open' );
            }
        });

        /*  [ Accordion ]
        - - - - - - - - - - - - - - - - - - - - */
        $( '.accordion.first-open > li:first-child' ).addClass( 'open' );
        $( '.accordion-title a' ).on( 'click', function(event) {
        	event.preventDefault();
            if( $( this ).parents( 'li' ).hasClass( 'open' ) ) {
            	$( this ).parents( 'li' ).removeClass( 'open' ).find( '.accordion-content' ).slideUp(400);
            } else {
                $( this ).parents( '.accordion' ).find( '.accordion-content' ).not( $( this ).parents( 'li' ).find( '.accordion-content' ) ).slideUp(400);
                $( this ).parents( '.accordion' ).find( '> li' ).not( $( this ).parents( 'li' ) ).removeClass( 'open' );
                $( this ).parents( 'li' ).addClass( 'open' ).find( '.accordion-content' ).slideDown(400);
            }
        });

        /*  [ Sticky Header ]
         - - - - - - - - - - - - - - - - - - - - */
        $( '.site-header' ).sticky({ topSpacing: 0 });

        /*  [ Testimonials Slider ]
         - - - - - - - - - - - - - - - - - - - - */
        var owlPreview = $( '.testi-preview .owl-carousel' );
        owlPreview.owlCarousel({
            nav: false,
            dots: true,
            items: 1,
            responsive : {
                0 : {
                    items: 1
                }
            },
            mouseDrag: false,
            touchDrag: false,
            animateIn: 'bounceIn',
            animateOut: 'zoomOut'
        });
		
        /*  [ Twitter Slider ]
         - - - - - - - - - - - - - - - - - - - - */
        $( '.twitter-slider .owl-carousel' ).owlCarousel({
            nav: false,
            dots: true,
            items: 1,
            autoplay: true,
            responsive : {
                0 : {
                    items: 1
                }
            }
        });

        /*  [ Blog Slider ]
         - - - - - - - - - - - - - - - - - - - - */
        $( '.blog-slider' ).owlCarousel({
            nav: false,
            dots: true,
            items: 3,
            margin: 30,
            loop: true,
            autoplay: true,
            responsive : {
                0 : {
                    items: 1
                },
                541 : {
                    items: 2
                },
                768 : {
                    items: 2
                },
                990 : {
                    items: 3
                }
            }
        });

        /*  [ Services Slider ]
         - - - - - - - - - - - - - - - - - - - - */
        $( '.services-slider .owl-carousel' ).owlCarousel({
            nav: false,
            dots: true,
            items: 4,
            margin: 20,
            loop: true,
            autoplay: true,
            responsive : {
                0 : {
                    items: 1
                },
                541 : {
                    items: 2
                },
                980 : {
                    items: 3
                },
                1200 : {
                    items: 4
                }
            }
        });

        /*  [ Counter Up ]
         - - - - - - - - - - - - - - - - - - - - */
        $( '.counter-number' ).counterUp({
            delay: 10,
        });

        /*  [ Back to top ]
        - - - - - - - - - - - - - - - - - - - - */
        $('.back-to-top').on( 'click', function(e) {
            e.preventDefault();
            $("html, body").animate({
                scrollTop: 0
            }, 700);
        });
	});
})(jQuery);