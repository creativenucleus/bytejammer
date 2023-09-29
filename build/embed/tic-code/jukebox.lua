SIN,ABS=math.sin,math.abs
PLAYLIST_ITEM_COUNT=--[[$PLAYLIST_ITEM_COUNT]]--

T=0
Y_TITLE=0
function BDR(y)
	if y>Y_TITLE and y<Y_TITLE+40 then
		for i=1,14 do
			local addr=0x3fc0+i*3
			local r=.6+SIN(i+y*.1+T*.04)*.4
			local g=.6+SIN(1+i*.8+y*.13+T*.03)*.4
			local b=.6+SIN(2+i*1.2+y*.07+T*.05)*.4
			poke(addr+0,r*255)
			poke(addr+1,g*255)
			poke(addr+2,b*255)
		end
	end
end

function BOOT()
	local addr=0x3fc0+15*3
	poke(addr+0,255)
	poke(addr+1,255)
	poke(addr+2,255)
end

function TIC()
	poke(0x3ffb,0)
	cls(0)

	Y_TITLE=10+SIN(T*.1)*4
	local title="TICJAMMER"
	local x=37
	for i=1,#title do
		local y=Y_TITLE+15-ABS(SIN(i*.4+T*.1))*10
		x=x+print(title:sub(i,i),x,y,i,false,3)
	end
  
	centrePrint("(--[[RELEASE_TITLE]]-- edition)",50,15)
	centrePrint("Jukebox Mode: " .. PLAYLIST_ITEM_COUNT .. " items",60,15)
  
    local texts={
 	    "Thanks",
		"TIC/Battle version: Nesbox, rho, Superogue",
		"Prior art: Aldroid, Gasman",
		"LCDZ: Totetmatt, PSEnough",
		"Additional help: Mantratronic, Violet Procyon",
		"Nusan, and the Field-FX community",
    }
 
    for t=1,#texts do
        local y=65+t*9
        local c=1+(14-t)%15
		centrePrint(texts[t],y,c)
    end
    T=T+1
end

function centrePrint(text,y,c)
	w=print(text,240,0,0)
	print(text,120-w/2,y,c)
end