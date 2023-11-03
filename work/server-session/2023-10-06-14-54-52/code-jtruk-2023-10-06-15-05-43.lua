function TIC()
	local t=time()/10000
	for x=-120,119 do
		for y=-68,67 do
   local a=math.atan2(y,x)/(math.pi*2)+t
			local d=(x^2+y^2)^.5
   local c=1+(a*15+d/100)%15
			pix(120+x,68+y,c)
		end
	end 
end
