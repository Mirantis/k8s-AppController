/*!
Copyright (c) The Cytoscape Consortium

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the “Software”), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

;(function(){ 'use strict';

  var $ = typeof jQuery === typeof undefined ? null : jQuery;

  var defaults = {
    menuRadius: 100, // the radius of the circular menu in pixels
    selector: 'node', // elements matching this Cytoscape.js selector will trigger cxtmenus
    commands: [ // an array of commands to list in the menu or a function that returns the array
      /*
      { // example command
        fillColor: 'rgba(200, 200, 200, 0.75)', // optional: custom background color for item
        content: 'a command name' // html/text content to be displayed in the menu
        select: function(ele){ // a function to execute when the command is selected
          console.log( ele.id() ) // `ele` holds the reference to the active element
        }
      }
      */
    ], // function( ele ){ return [ /*...*/ ] }, // example function for commands
    fillColor: 'rgba(0, 0, 0, 0.75)', // the background colour of the menu
    activeFillColor: 'rgba(92, 194, 237, 0.75)', // the colour used to indicate the selected command
    activePadding: 20, // additional size in pixels for the active command
    indicatorSize: 24, // the size in pixels of the pointer to the active command
    separatorWidth: 3, // the empty spacing in pixels between successive commands
    spotlightPadding: 4, // extra spacing in pixels between the element and the spotlight
    minSpotlightRadius: 24, // the minimum radius in pixels of the spotlight
    maxSpotlightRadius: 38, // the maximum radius in pixels of the spotlight
    openMenuEvents: 'cxttapstart taphold', // cytoscape events that will open the menu (space separated)
    itemColor: 'white', // the colour of text in the command's content
    itemTextShadowColor: 'black', // the text shadow colour of the command's content
    zIndex: 9999, // the z-index of the ui div
    atMouse: false // draw menu at mouse position
  };

  // registers the extension on a cytoscape lib ref
  var register = function( cytoscape, $ ){
    if( !cytoscape ){ return; } // can't register if cytoscape unspecified

    cytoscape('core', 'cxtmenu', function(params){
      var options = $.extend(true, {}, defaults, params);
      var fn = params;
      var cy = this;
      var $container = $( cy.container() );
      var target;

      function getOffset( $ele ){
        var offset = $ele.offset();

        offset.left += parseFloat( $ele.css('padding-left') );
        offset.left += parseFloat( $ele.css('border-left-width') );

        offset.top += parseFloat( $ele.css('padding-top') );
        offset.top += parseFloat( $ele.css('border-top-width') );

        return offset;
      }

      var data = {
        options: options,
        handlers: []
      };
      var $wrapper = $('<div class="cxtmenu"></div>'); data.$container = $wrapper;
      var $parent = $('<div></div>');
      var $canvas = $('<canvas></canvas>');
      var commands = [];
      var c2d = $canvas[0].getContext('2d');
      var r = options.menuRadius;
      var containerSize = (r + options.activePadding)*2;
      var activeCommandI = undefined;
      var offset;

      $container.prepend( $wrapper );
      $wrapper.append( $parent );
      $parent.append( $canvas );

      $wrapper.css({
        position: 'absolute',
        zIndex: options.zIndex
      });

      $parent.css({
        width: containerSize + 'px',
        height: containerSize + 'px',
        position: 'absolute',
        zIndex: 1,
        marginLeft: - options.activePadding + 'px',
        marginTop: - options.activePadding + 'px'
      }).hide();

      $canvas[0].width = containerSize;
      $canvas[0].height = containerSize;

    function createMenuItems() {
      $('.cxtmenu-item').remove();
      var dtheta = 2 * Math.PI / (commands.length);
      var theta1 = Math.PI / 2;
      var theta2 = theta1 + dtheta;

      for (var i = 0; i < commands.length; i++) {
        var command = commands[i];

        var midtheta = (theta1 + theta2) / 2;
        var rx1 = 0.66 * r * Math.cos(midtheta);
        var ry1 = 0.66 * r * Math.sin(midtheta);

        var $item = $('<div class="cxtmenu-item"></div>');
        $item.css({
          color: options.itemColor,
          cursor: 'default',
          display: 'table',
          'text-align': 'center',
          //background: 'red',
          position: 'absolute',
          'text-shadow': '-1px -1px ' + options.itemTextShadowColor + ', 1px -1px ' + options.itemTextShadowColor + ', -1px 1px ' + options.itemTextShadowColor + ', 1px 1px ' + options.itemTextShadowColor,
          left: '50%',
          top: '50%',
          'min-height': r * 0.66,
          width: r * 0.66,
          height: r * 0.66,
          marginLeft: rx1 - r * 0.33,
          marginTop: -ry1 - r * 0.33
        });

        var $content = $('<div class="cxtmenu-content">' + command.content + '</div>');
        $content.css({
          'width': r * 0.66,
          'height': r * 0.66,
          'vertical-align': 'middle',
          'display': 'table-cell'
        });

        if (command.disabled) {
          $content.addClass('cxtmenu-disabled');
        }

        $parent.append($item);
        $item.append($content);


        theta1 += dtheta;
        theta2 += dtheta;
      }
    }

      function queueDrawBg( rspotlight ){
        redrawQueue.drawBg = [ rspotlight ];
      }

      function drawBg( rspotlight ){
        rspotlight = rspotlight !== undefined ? rspotlight : rs;

        c2d.globalCompositeOperation = 'source-over';

        c2d.clearRect(0, 0, containerSize, containerSize);

        // draw background items
        c2d.fillStyle = options.fillColor;
        var dtheta = 2*Math.PI/(commands.length);
        var theta1 = Math.PI/2;
        var theta2 = theta1 + dtheta;

        for( var index = 0; index < commands.length; index++ ){
          var command = commands[index];

          if( command.fillColor ){
            c2d.fillStyle = command.fillColor;
          }
          c2d.beginPath();
          c2d.moveTo(r + options.activePadding, r + options.activePadding);
          c2d.arc(r + options.activePadding, r + options.activePadding, r, 2*Math.PI - theta1, 2*Math.PI - theta2, true);
          c2d.closePath();
          c2d.fill();

          theta1 += dtheta;
          theta2 += dtheta;

          c2d.fillStyle = options.fillColor;
        }

        // draw separators between items
        c2d.globalCompositeOperation = 'destination-out';
        c2d.strokeStyle = 'white';
        c2d.lineWidth = options.separatorWidth;
        theta1 = Math.PI/2;
        theta2 = theta1 + dtheta;

        for( var i = 0; i < commands.length; i++ ){
          var rx1 = r * Math.cos(theta1);
          var ry1 = r * Math.sin(theta1);
          c2d.beginPath();
          c2d.moveTo(r + options.activePadding, r + options.activePadding);
          c2d.lineTo(r + options.activePadding + rx1, r + options.activePadding - ry1);
          c2d.closePath();
          c2d.stroke();

          theta1 += dtheta;
          theta2 += dtheta;
        }


        c2d.fillStyle = 'white';
        c2d.globalCompositeOperation = 'destination-out';
        c2d.beginPath();
        c2d.arc(r + options.activePadding, r + options.activePadding, rspotlight + options.spotlightPadding, 0, Math.PI*2, true);
        c2d.closePath();
        c2d.fill();

        c2d.globalCompositeOperation = 'source-over';
      }

      function queueDrawCommands( rx, ry, theta ){
        redrawQueue.drawCommands = [ rx, ry, theta ];
      }

      function drawCommands( rx, ry, theta ){
        var dtheta = 2*Math.PI/(commands.length);
        var theta1 = Math.PI/2;
        var theta2 = theta1 + dtheta;

        theta1 += dtheta * activeCommandI;
        theta2 += dtheta * activeCommandI;

        c2d.fillStyle = options.activeFillColor;
        c2d.strokeStyle = 'black';
        c2d.lineWidth = 1;
        c2d.beginPath();
        c2d.moveTo(r + options.activePadding, r + options.activePadding);
        c2d.arc(r + options.activePadding, r + options.activePadding, r + options.activePadding, 2*Math.PI - theta1, 2*Math.PI - theta2, true);
        c2d.closePath();
        c2d.fill();

        c2d.fillStyle = 'white';
        c2d.globalCompositeOperation = 'destination-out';

        // clear the indicator
        c2d.beginPath();
        c2d.translate( r + options.activePadding + rx/r*(rs + options.spotlightPadding - options.indicatorSize/4), r + options.activePadding + ry/r*(rs + options.spotlightPadding - options.indicatorSize/4) );
        c2d.rotate( Math.PI/4 - theta );
        c2d.fillRect(-options.indicatorSize/2, -options.indicatorSize/2, options.indicatorSize, options.indicatorSize);
        c2d.closePath();
        c2d.fill();

        c2d.setTransform(1, 0, 0, 1, 0, 0);

        // clear the spotlight
        c2d.beginPath();
        c2d.arc(r + options.activePadding, r + options.activePadding, rs + options.spotlightPadding, 0, Math.PI*2, true);
        c2d.closePath();
        c2d.fill();

        c2d.globalCompositeOperation = 'source-over';
      }

      var redrawing = true;
      var redrawQueue = {};
      var raf = window.requestAnimationFrame || window.webkitRequestAnimationFrame || window.mozRequestAnimationFrame;
      var redraw = function(){
        if( redrawQueue.drawBg ){
          drawBg.apply( null, redrawQueue.drawBg );
        }

        if( redrawQueue.drawCommands ){
          drawCommands.apply( null, redrawQueue.drawCommands );
        }

        redrawQueue = {};

        if( redrawing ){
          raf( redraw );
        }
      };

      redraw(); // kick off

      var ctrx, ctry, rs;

      var bindings = {
        on: function(events, selector, fn){

          var _fn = fn;
          if( selector === 'core'){
            _fn = function( e ){
              if( e.cyTarget === cy ){ // only if event target is directly core
                return fn.apply( this, [ e ] );
              }
            };
          }

          data.handlers.push({
            events: events,
            selector: selector,
            fn: _fn
          });

          if( selector === 'core' ){
            cy.on(events, _fn);
          } else {
            cy.on(events, selector, _fn);
          }

          return this;
        }
      };

      function addEventListeners(){
        var grabbable;
        var inGesture = false;
        var dragHandler;
        var zoomEnabled;
        var panEnabled;

        var restoreZoom = function(){
          if( zoomEnabled ){
            cy.userZoomingEnabled( true );
          }
        };

        var restoreGrab = function(){
          if( grabbable ){
            target.grabify();
          }
        };

        var restorePan = function(){
          if( panEnabled ){
            cy.userPanningEnabled( true );
          }
        };

        bindings
          .on(options.openMenuEvents, options.selector, function(e){
            target = this; // Remember which node the context menu is for
            var ele = this;
            var isCy = this === cy;

            if( typeof options.commands === 'function' ){
              commands = options.commands(target);
            } else {
              commands = options.commands;
            }

            if( !commands || commands.length == 0 ){ return; }

            zoomEnabled = cy.userZoomingEnabled();
            cy.userZoomingEnabled( false );

            panEnabled = cy.userPanningEnabled();
            cy.userPanningEnabled( false );

            grabbable = target.grabbable &&  target.grabbable();
            if( grabbable ){
              target.ungrabify();
            }

            var rp, rw, rh;
            if( !isCy && ele.isNode() && !options.atMouse ){
              rp = ele.renderedPosition();
              rw = ele.renderedWidth();
              rh = ele.renderedHeight();
            } else {
              rp = e.cyRenderedPosition;
              rw = 1;
              rh = 1;
            }

            offset = getOffset( $container );

            ctrx = rp.x;
            ctry = rp.y;

            createMenuItems();

            $parent.show().css({
              'left': rp.x - r + 'px',
              'top': rp.y - r + 'px'
            });

            rs = Math.max(rw, rh)/2;
            rs = Math.max(rs, options.minSpotlightRadius);
            rs = Math.min(rs, options.maxSpotlightRadius);

            queueDrawBg();

            activeCommandI = undefined;

            inGesture = true;
          })

          .on('cxtdrag tapdrag', options.selector, dragHandler = function(e){

            if( !inGesture ){ return; }

            var origE = e.originalEvent;
            var isTouch = origE.touches && origE.touches.length > 0;

            var pageX = isTouch ? origE.touches[0].pageX : origE.pageX;
            var pageY = isTouch ? origE.touches[0].pageY : origE.pageY;

            activeCommandI = undefined;

            var dx = pageX - offset.left - ctrx;
            var dy = pageY - offset.top - ctry;

            if( dx === 0 ){ dx = 0.01; }

            var d = Math.sqrt( dx*dx + dy*dy );
            var cosTheta = (dy*dy - d*d - dx*dx)/(-2 * d * dx);
            var theta = Math.acos( cosTheta );

            if( d < rs + options.spotlightPadding ){
              queueDrawBg();
              return;
            }

            queueDrawBg();

            var rx = dx*r / d;
            var ry = dy*r / d;

            if( dy > 0 ){
              theta = Math.PI + Math.abs(theta - Math.PI);
            }

            var dtheta = 2*Math.PI/(commands.length);
            var theta1 = Math.PI/2;
            var theta2 = theta1 + dtheta;

            for( var i = 0; i < commands.length; i++ ){
              var command = commands[i];

              var inThisCommand = theta1 <= theta && theta <= theta2
                || theta1 <= theta + 2*Math.PI && theta + 2*Math.PI <= theta2;

              if( command.disabled ){
                inThisCommand = false;
              }

              if( inThisCommand ){
                activeCommandI = i;
                break;
              }

              theta1 += dtheta;
              theta2 += dtheta;
            }

            queueDrawCommands( rx, ry, theta );
          })

          .on('tapdrag', dragHandler)

          .on('cxttapend tapend', options.selector, function(e){
            var ele = this;
            $parent.hide();

            if( activeCommandI !== undefined ){
              var select = commands[ activeCommandI ].select;

              if( select ){
                select.apply( ele, [ele] );
                activeCommandI = undefined;
              }
            }

            inGesture = false;

            restoreGrab();
            restoreZoom();
            restorePan();
          })

          .on('cxttapend tapend', function(e){
            $parent.hide();

            inGesture = false;

            restoreGrab();
            restoreZoom();
            restorePan();
          })
        ;
      }

      function removeEventListeners(){
        var handlers = data.handlers;

        for( var i = 0; i < handlers.length; i++ ){
          var h = handlers[i];

          if( h.selector === 'core' ){
            cy.off(h.events, h.fn);
          } else {
            cy.off(h.events, h.selector, h.fn);
          }
        }
      }

      function destroyInstance(){
        redrawing = false;

        removeEventListeners();

        $wrapper.remove();
      }

      addEventListeners();

      return {
        destroy: function(){
          destroyInstance();
        }
      };

    });

  }; // reg

  if( typeof module !== 'undefined' && module.exports ){ // expose as a commonjs module
    module.exports = register;
  }

  if( typeof define !== 'undefined' && define.amd ){ // expose as an amd/requirejs module
    define('cytoscape-cxtmenu', function(){
      return register;
    });
  }

  if( typeof cytoscape !== typeof undefined && $ ){ // expose to global cytoscape (i.e. window.cytoscape)
    register( cytoscape, $ );
  }

})();
