function TIC()
	local t=time()/1000
	for x=-120,119 do
		for y=-68,67 do
   local a=.5+(math.atan2(x,y)/math.pi*2)*.5+t
			local d=(x^2+y^2)^.5
   local c=1*a
			pix(120+x,68+y,c)
		end
	end 
end
